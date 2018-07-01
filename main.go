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

const defaultTomlConfig = `
title = "Tailon file viewer"
relative-root = "/"
listen-addr = ":8080"
allowed-commands = ["tail", "grep", "sed", "awk"]

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

const scriptDescription = `
Usage: tailon [options] -c <config file>
Usage: tailon [options] <filespec> [<filespec> ...]

Tailon is a webapp for looking at and searching through log files.
`

const scriptEpilog = `
The format of the one or more positional 'filespec' arguments is:
  [[glob|dir|file],alias=name,group=name,]<path>

Example usage:
  tailon /var/log/messages /var/log/debug /var/log/*.log
  tailon -b localhost:8080 -c config.ini
`

const configFileHelp = `
<todo>
`

type CommandSpec struct {
	Stdin   string
	Action  []string
	Default string
}

func parseTomlConfig(config string) (*toml.Tree, map[string]CommandSpec, error) {
	cfg, err := toml.Load(config)

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

	return cfg, commands, err
}

type FileSpec struct {
	Path  string
	Type  string
	Alias string
	Group string
}

func parseFileSpec(spec string) (FileSpec, error) {
	var filespec FileSpec
	parts := strings.Split(spec, ",")

	if length := len(parts); length == 1 {
		return FileSpec{spec, "file", "", ""}, nil
	}

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
	RelativeRoot        string
	BindAddr            string
	ConfigPath          string
	WrapLinesInitial    bool
	TailLinesInitial    int
	AllowedCommandNames []string

	CommandSpecs   map[string]CommandSpec
	CommandScripts map[string]string
	FileSpecs      []FileSpec
}

var config = Config{}

func main() {
	defaultConfig, commandSpecs, _ := parseTomlConfig(defaultTomlConfig)

	printHelp := flag.BoolP("help", "h", false, "Show this help message and exit")
	printConfigHelp := flag.BoolP("help-config", "e", false, "Show config file help and exit")

	flag.StringVarP(&config.BindAddr, "bind", "b", defaultConfig.Get("listen-addr").(string), "Listen on the specified address and port")
	flag.StringVarP(&config.ConfigPath, "config", "c", "", "")
	flag.StringVarP(&config.RelativeRoot, "relative-root", "r", defaultConfig.Get("relative-root").(string), "webapp relative root")

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

	config.RelativeRoot = "/" + strings.TrimLeft(config.RelativeRoot, "/")
	config.RelativeRoot = strings.TrimRight(config.RelativeRoot, "/") + "/"
	mapstructure.Decode(defaultConfig.Get("allowed-commands"), &config.AllowedCommandNames)
	config.CommandSpecs = commandSpecs

	config.CommandScripts = make(map[string]string)
	for cmd, values := range config.CommandSpecs {
		config.CommandScripts[cmd] = values.Default
	}

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting", config.RelativeRoot)

	server := SetupServer(config, logger)
	server.ListenAndServe()
}
