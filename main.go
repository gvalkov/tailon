// A webapp for looking at and searching through files.
package main

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml"
	flag "github.com/spf13/pflag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

const scriptDescription = `
Usage: tailon -c <config file>
Usage: tailon [options] <filespec> [<filespec> ...]

Tailon is a webapp for searching through files and streams.
`

const scriptEpilog = `
Tailon can be configured via a TOML config file or command-line flags.

The command-line interface expects one or more <filespec> arguments, which
specify the files to serve. The format is:

  [alias=name,group=name]<source>

The "source" specifier can be a file name, glob or directory. The optional
"alias=" and "group=" specifiers change the display name of files in the UI
and the group in which they appear. 

A file specifier points to a single, possibly non-existent file. The file
name can be overwritten with "alias=". For example:

  tailon alias=error.log,/var/log/apache/error.log

A glob evaluates to the list of files that match a shell filename pattern.
The pattern is evaluated each time the file list is refreshed. An "alias="
specifier overwrites the parent directory of each matched file in the UI.

  tailon "/var/log/apache/*.log" "alias=nginx,/var/log/nginx/*.log"

If a directory is given, all files under it are served recursively.

  tailon /var/log/apache/ /var/log/nginx/

Example usage:
  tailon file1.txt file2.txt file3.txt
  tailon alias=messages,/var/log/messages "/var/log/*.log"
  tailon -b localhost:8080,localhost:8081 -c config.toml

See "--help-config" for configuration file usage.
`

const configFileHelp = `
The following options can be set through the config file:

  # The <title> element of the of the webapp.
  title = "Tailon file viewer"

  # The root of the web application.
  relative-root = "/"

  # The addresses to listen on. Can be an address:port combination or an unix socket.
  listen-addr = [":8080"]

  # Allow downloading of known files (i.e those matched by a filespec).
  allow-download = true

  # Commands that will appear in the UI.
  allow-commands = ["tail", "grep", "sed", "awk"]

  # A table of commands that the backend can execute. This is best illustrated by
  # the default configuration listed below.
  [commands]

  # File, glob and dir filespecs are similar in principle to their
  # command-line counterparts.

At startup tailon loads the following default configuration:
`

const defaultTomlConfig = `
  title = "Tailon file viewer"
  relative-root = "/"
  listen-addr = [":8080"]
  allow-download = true
  allow-commands = ["tail", "grep", "sed", "awk"]

  [commands]

    [commands.tail]
    action = ["tail", "-n", "$lines", "-F", "$path"]

    [commands.grep]
    stdin = "tail"
    action = ["grep", "--text", "--line-buffered", "--color=never", "-e", "$script"]
    default = ".*"

    [commands.sed]
    stdin = "tail"
    action = ["sed", "-u", "-e", "$script"]
    default = "s/.*/&/"

    [commands.awk]
    stdin = "tail"
    action = ["awk", "--sandbox", "$script"]
    default = "{print $0; fflush()}"
`

// CommandSpec defines a command that the server can execute.
type CommandSpec struct {
	Stdin   string
	Action  []string
	Default string
}

func parseTomlConfig(config string) (*toml.Tree, map[string]CommandSpec) {
	cfg, err := toml.Load(config)
	if err != nil {
		log.Fatal("Error parsing config: ", err)
	}

	commands := make(map[string]CommandSpec)

	cfgCommands := cfg.Get("commands").(*toml.Tree).ToMap()
	for key, value := range cfgCommands {
		command := CommandSpec{}
		err := mapstructure.Decode(value, &command)
		if err != nil {
			log.Fatal(err)
		}
		commands[key] = command
	}

	return cfg, commands
}

// FileSpec is an instance of a file to be monitored. These are mapped to
// os.Args or the [files] elements in the config file.
type FileSpec struct {
	Path  string
	Type  string
	Alias string
	Group string
}

// Parse a string into a filespec. Example inputs are:
//   alias=1,group=2,/var/log/messages
//   /var/log/
//   /var/log/*
func parseFileSpec(spec string) (FileSpec, error) {
	var filespec FileSpec
	var path string
	parts := strings.Split(spec, ",")

	if length := len(parts); length == 1 {
		path = parts[0]
	} else {
		// The last part is the path. We'll probably need a more robust
		// solution in the future.
		path, parts = parts[len(parts)-1], parts[:len(parts)-1]
	}

	if strings.ContainsAny(path, "*?[]") {
		filespec.Type = "glob"
	} else {
		stat, err := os.Lstat(path)
		if os.IsNotExist(err) || stat.Mode().IsRegular() {
			filespec.Type = "file"
		} else if stat.Mode().IsDir() {
			filespec.Type = "dir"
		}
	}

	for _, part := range parts {
		if strings.HasPrefix(part, "group=") {
			group := strings.SplitN(part, "=", 2)[1]
			group = strings.Trim(group, "'\" ")
			filespec.Group = group
		} else if strings.HasPrefix(part, "alias=") {
			filespec.Alias = strings.SplitN(part, "=", 2)[1]
		}
	}

	if filespec.Type == "" {
		filespec.Type = "file"
	}
	filespec.Path = path
	return filespec, nil

}

// Config contains all backend and frontend configuration options and relevant state.
type Config struct {
	RelativeRoot      string
	BindAddr          []string
	ConfigPath        string
	WrapLinesInitial  bool
	TailLinesInitial  int
	AllowCommandNames []string
	AllowDownload     bool

	CommandSpecs   map[string]CommandSpec
	CommandScripts map[string]string
	FileSpecs      []FileSpec
}

func makeConfig(configContent string) *Config {
	defaults, commandSpecs := parseTomlConfig(configContent)

	// Convert the list of bind addresses from []interface{} to []string.
	addrsA := defaults.Get("listen-addr").([]interface{})
	addrsB := make([]string, len(addrsA))
	for i := range addrsA {
		addrsB[i] = addrsA[i].(string)
	}

	config := Config{
		BindAddr:      addrsB,
		RelativeRoot:  defaults.Get("relative-root").(string),
		AllowDownload: defaults.Get("allow-download").(bool),
		CommandSpecs:  commandSpecs,
	}

	mapstructure.Decode(defaults.Get("allow-commands"), &config.AllowCommandNames)
	return &config
}

var config = &Config{}

func main() {
	config = makeConfig(defaultTomlConfig)

	printHelp := flag.BoolP("help", "h", false, "Show this help message and exit")
	printConfigHelp := flag.BoolP("help-config", "e", false, "Show configuration file help and exit")
	bindAddr := flag.StringP("bind", "b", strings.Join(config.BindAddr, ","), "Address and port to listen on")

	flag.StringVarP(&config.RelativeRoot, "relative-root", "r", config.RelativeRoot, "Webapp relative root")
	flag.BoolVarP(&config.AllowDownload, "allow-download", "a", config.AllowDownload, "Allow file downloads")
	flag.StringVarP(&config.ConfigPath, "config", "c", "", "Path to TOML configuration file")
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, strings.TrimLeft(scriptDescription, "\n"))
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, strings.TrimRight(scriptEpilog, "\n"))
		os.Exit(0)
	}

	if *printHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *printConfigHelp {
		fmt.Fprintf(os.Stderr, "%s\n\n%s\n", strings.Trim(configFileHelp, "\n"), strings.Trim(defaultTomlConfig, "\n"))
		os.Exit(0)
	}

	config.BindAddr = strings.Split(*bindAddr, ",")

	// If a configuration file is specified, read all options from it. This discards all options set on the command-line.
	if config.ConfigPath != "" {
		if b, err := ioutil.ReadFile(config.ConfigPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading config file '%s': %s\n", config.ConfigPath, err)
			os.Exit(1)
		} else {
			config = makeConfig(string(b))
		}
	}

	// Ensure that relative root is always '/' or '/$arg/'.
	config.RelativeRoot = "/" + strings.TrimLeft(config.RelativeRoot, "/")
	config.RelativeRoot = strings.TrimRight(config.RelativeRoot, "/") + "/"

	// Handle command-line file specs.
	filespecs := make([]FileSpec, len(flag.Args()))
	for _, spec := range flag.Args() {
		if filespec, err := parseFileSpec(spec); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing argument '%s': %s\n", spec, err)
			os.Exit(1)
		} else {
			filespecs = append(filespecs, filespec)
		}
	}
	config.FileSpecs = filespecs

	if len(config.FileSpecs) == 0 {
		fmt.Fprintln(os.Stderr, "No files specified on command-line or in config file")
		os.Exit(2)
	}

	config.CommandScripts = make(map[string]string)
	for cmd, values := range config.CommandSpecs {
		config.CommandScripts[cmd] = values.Default
	}

	log.Print("Generate initial file listing")
	createListing(config.FileSpecs)

	var wg sync.WaitGroup
	for _, addr := range config.BindAddr {
		wg.Add(1)
		go startServer(config, addr)
	}
	wg.Wait()

}

func startServer(config *Config, bindAddr string) {
	loggerHTML := log.New(os.Stdout, "", log.LstdFlags)
	loggerHTML.Printf("Server start, relative-root: %s, bind-addr: %s\n", config.RelativeRoot, bindAddr)

	server := setupServer(config, bindAddr, loggerHTML)

	if strings.Contains(bindAddr, ":") {
		server.ListenAndServe()
	} else {
		os.Remove(bindAddr)

		unixAddr, _ := net.ResolveUnixAddr("unix", bindAddr)
		unixListener, err := net.ListenUnix("unix", unixAddr)
		unixListener.SetUnlinkOnClose(true)

		if err != nil {
			panic(err)
		}

		defer unixListener.Close()
		server.Serve(unixListener)
	}
}
