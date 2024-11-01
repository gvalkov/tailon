<p align="center">
  <img src="https://user-images.githubusercontent.com/190997/42879022-6d5915fc-8a8f-11e8-8fe6-903c06bd52a9.png?raw=True" width="450px">
</p>

# Tailon

[![GoDoc](https://godoc.org/github.com/gvalkov/tailon?status.svg)](https://godoc.org/github.com/gvalkov/tailon)
[![Go Report Card](https://goreportcard.com/badge/github.com/gvalkov/tailon)](https://goreportcard.com/report/github.com/gvalkov/tailon)
[![Apache License](https://img.shields.io/badge/license-Apache-blue.svg)](https://github.com/gvalkov/tailon/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/gvalkov/tailon.svg)](https://github.com/gvalkov/tailon/releases)

Tailon is a webapp for looking at and searching through files and streams. In a
nutshell, it is a web wrapper around the following commands:

```
tail -f
tail -f | grep
tail -f | awk
tail -f | sed
```

## Install

Download a build for your platform from the [releases] page or install using `go get`:

```
go get -u github.com/gvalkov/tailon
```

A container image is also available:

```
docker run --rm gvalkov/tailon --help
```

## Usage

Tailon is a webapp that streams the output of commands such as `tail` and
`grep`. It can be configured with command-line flags or with a [toml] config
file. Some options, like adding new commands, are only available through the
configuration file.

To get started, run tailon with the list of files that you wish to monitor.

```
tailon /var/log/apache/access.log /var/log/apache/error.log /var/log/messages
```

Tailon can serve single files, globs or whole directory trees. Tailon’s
server-side functionality is summarized entirely in its help message:

[//]: # (run "make README.md" to update the next section with the output of tailon --help)

[//]: # (BEGIN HELP_USAGE)
```
Usage: tailon -c <config file>
Usage: tailon [options] <filespec> [<filespec> ...]

Tailon is a webapp for searching through files and streams.

  -a, --allow-download         Allow file downloads (default true)
  -b, --bind string            Address and port to listen on (default ":8080")
  -c, --config string          Path to TOML configuration file
  -h, --help                   Show this help message and exit
  -e, --help-config            Show configuration file help and exit
  -r, --relative-root string   Webapp relative root (default "/")

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
```
[//]: # (END HELP_USAGE)

[//]: # (BEGIN HELP_CONFIG)
```
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
```
[//]: # (END HELP_CONFIG)

## Security

Tailon runs commands on the server it is installed on. While commands that
accept a script argument (such as awk, sed and grep) should be invulnerable
to shell injection, they may still allow for arbitrary command execution
and unrestricted access to the filesystem.

To clarify this point, consider the following script input to the `sed` command:

```
s/a/b'; cat /etc/secrets
```

This will result in an error, as tailon does not invoke commands through a
shell. On the other hand, the following command is a perfectly valid sed script
that has the same effect as the above attempt for shell injection:

```
r /etc/secrets
```

The default set of enabled commands (tail, grep and awk) should be safe to use.
GNU awk is run in [sandbox] mode, which prevents scripts from accessing your
system, either through the `system()` builtin or by using input redirection.

By default, tailon is accessible to anyone who knows the server address and
port.


## Development

### Frontend

To work on the frontend, make sure you're building with the `dev` build tag:

```
go build -tags dev
```

This will ensure that the `tailon` binary is reading assets from the
`frontend/dist` directory instead of from `frontend/assets_vfsdata.go`.
To compile the web assets, use `make all` in `frontend`.

The `make watch` goal can be used to continuously update the bundles as you
make changes to the sources.

Note that the frontend bundles are committed in order to make life easier for
people that only want to work on the backend.

### Backend

The backend is written in straightforward go that tries to do as much as
possible using only the standard library.


### TODO

* Directory serving.

* User-specified TOML configuration files.

* Basic and digest authentication.

* Add a 'command' filespec - e.g. `"command,journalctl -u nginx"`.

* Better configuration dialog.

* Add interface themes - e.g. light, dark and solarized.

* Add ability to change font family and size.

* Windows support (can use one of the Go tail implementations).

* Implement [wtee].

* Stderr is streamed to the client, but it is not handled at the moment.

* Support ANSI color codes.


### Testing

The project has unit-tests, which you can run with `go test` and integration
tests which you can run with `cd tests; pytest`. Alternatively, you can run both
with `make test test-int`.

The integration tests are written in Python and use `pytest` and `aiohttp` to
interact with a running `tailon` instance.

```shell
make test/.venv  # run once to set up virtualenv and dependencies
make test-int
```


## What about the other tailon project?

This project is a full rewrite of the original [tailon] with the following goals in mind:

* Reduce maintenance overhead (especially on the frontend).
* Remove unwanted features and fix poor design choices.
* Learn more about Go and Vue.js.

In terms of tech, the following has changed:

* Backend from Python+Tornado to Go.
* Frontend from a very-custom Typescript solution to a simple ES5 + Vue.js app.
* Simplified asset pipeline (a short Makefile).
* Config file is now toml based.
* Fully self-contained executable.


## Similar Projects

* [clarity]
* [errorlog]
* [log.io]
* [rtail]
* [tailon]


Attributions
------------

Tailon's favicon was created from [this icon].


## License

Tailon is released under the terms of the [Apache 2.0 License].



[clarity]:   https://github.com/tobi/clarity
[tailon]:    https://github.com/gvalkov/tailon-legacy
[wtee]:      https://github.com/gvalkov/wtee
[toml]:      https://github.com/toml-lang/toml
[releases]:  https://github.com/gvalkov/tailon-next/releases
[errorlog]:  http://www.psychogenic.com/en/products/Errorlog.php
[log.io]:    http://logio.org/
[rtail]:     http://rtail.org/
[this icon]: http://www.iconfinder.com/icondetails/15150/48/terminal_icon
[sandbox]:   http://www.gnu.org/software/gawk/manual/html_node/Options.html#index-g_t_0040code_007b_002dS_007d-option-277
[Apache 2.0 License]: LICENSE
