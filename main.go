// A webapp for looking at and searching through files.
package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml"
	flag "github.com/spf13/pflag"
	"log"
	"strings"
)

const scriptDescription = `
Usage: tailon [options] -c <config file>
Usage: tailon [options] <filespec> [<filespec> ...]

Tailon is a webapp for looking at and searching through files and streams.
`

const scriptEpilog = `
Tailon can be configured through a config file or with command-line flags.

The command-line interface expects one or more filespec arguments, which
specify the files or directories to be served. The expected format is:

  [[glob|dir|file],alias=name,group=name,]<path>

The default filespec is 'file' and points to a single, possibly non-existent
file. The file name in the UI can be overwritten with the 'alias=' specifier.

The 'glob' filespec evaluates to the list of files that match a shell file
name pattern. The pattern is evaluated each time the file list is refreshed.
The 'alias=' specifier overwrites the parent directory of each matched file.

The 'dir' specifier evaluates to all files in a directory.

The "group=" specifier sets the group in which files appear in the file
dropdown of the toolbar.

Example usage:
  tailon alias=messages,/var/log/messages "glob:/var/log/*.log"
  tailon -b localhost:8080 -c config.toml
`

const configFileHelp = `
<todo>
`

const defaultTomlConfig = `
title = "Tailon file viewer"
relative-root = "/"
listen-addr = ":8080"
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

	cfg_commands := cfg.Get("commands").(*toml.Tree).ToMap()
	for key, value := range cfg_commands {
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
//   file,alias=1,group=2,/var/log/messages
//   /var/log/messages
//   glob,/var/log/*
func parseFileSpec(spec string) (FileSpec, error) {
	var filespec FileSpec
	parts := strings.Split(spec, ",")

	// If no specifiers are given, default is file.
	if length := len(parts); length == 1 {
		return FileSpec{spec, "file", "", ""}, nil
	}

	// The last part is the path. We'll probably need a more robust
	// solution in the future.
	path, parts := parts[len(parts)-1], parts[:len(parts)-1]

	for _, part := range parts {
		if strings.HasPrefix(part, "group=") {
			group := strings.SplitN(part, "=", 2)[1]
			group = strings.Trim(group, "'\" ")
			filespec.Group = group
		} else if strings.HasPrefix(part, "alias=") {
			filespec.Alias = strings.SplitN(part, "=", 2)[1]
		} else {
			switch part {
			case "file", "dir", "glob":
				filespec.Type = part
			}
		}

	}

	if filespec.Type == "" {
		filespec.Type = "file"
	}
	filespec.Path = path
	return filespec, nil

}

type Config struct {
	RelativeRoot      string
	BindAddr          string
	ConfigPath        string
	WrapLinesInitial  bool
	TailLinesInitial  int
	AllowCommandNames []string
	AllowDownload     bool

	CommandSpecs   map[string]CommandSpec
	CommandScripts map[string]string
	FileSpecs      []FileSpec
}

func makeConfig() *Config {
	defaults, commandSpecs := parseTomlConfig(defaultTomlConfig)

	config := Config{
		BindAddr:      defaults.Get("listen-addr").(string),
		RelativeRoot:  defaults.Get("relative-root").(string),
		AllowDownload: defaults.Get("allow-download").(bool),
		CommandSpecs:  commandSpecs,
	}

	mapstructure.Decode(defaults.Get("allow-commands"), &config.AllowCommandNames)
	return &config
}

var logger *log.Logger
var config = &Config{}

func main() {
	config = makeConfig()
	logger = log.New(os.Stderr, "main: ", log.LstdFlags)

	printHelp := flag.BoolP("help", "h", false, "Show this help message and exit")
	printConfigHelp := flag.BoolP("help-config", "e", false, "Show config file help and exit")

	flag.StringVarP(&config.BindAddr, "bind", "b", config.BindAddr, "Listen on the specified address and port")
	flag.StringVarP(&config.ConfigPath, "config", "c", "", "")
	flag.StringVarP(&config.RelativeRoot, "relative-root", "r", config.RelativeRoot, "webapp relative root")
	flag.BoolVarP(&config.AllowDownload, "allow-download", "a", config.AllowDownload, "allow file downloads")
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, strings.TrimLeft(scriptDescription, "\n"))
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, strings.TrimRight(scriptEpilog, "\n"))
		os.Exit(2)
	}

	if *printHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *printConfigHelp {
		fmt.Fprintln(os.Stderr, strings.Trim(configFileHelp, "\n"))
		os.Exit(0)
	}

	// Ensure that relative root is always '/' or '/$arg/'.
	config.RelativeRoot = "/" + strings.TrimLeft(config.RelativeRoot, "/")
	config.RelativeRoot = strings.TrimRight(config.RelativeRoot, "/") + "/"

	// Handle command-line file specs
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

	logger.Print("Generate initial file listing")
	createListing(config.FileSpecs)

	loggerHtml := log.New(os.Stdout, "http: ", log.LstdFlags)
	loggerHtml.Printf("Server start, relative-root: %s, bind-addr: %s\n", config.RelativeRoot, config.BindAddr)

	server := SetupServer(config, loggerHtml)
	server.ListenAndServe()
}
