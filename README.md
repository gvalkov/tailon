<p align="center">
  <img src="https://user-images.githubusercontent.com/190997/42879022-6d5915fc-8a8f-11e8-8fe6-903c06bd52a9.png?raw=True" width="450px">
</p>

# Tailon [![Build Status](https://travis-ci.com/gvalkov/tailon.svg?branch=master)](https://travis-ci.com/gvalkov/tailon) [![GoDoc](https://godoc.org/github.com/gvalkov/tailon?status.svg)](https://godoc.org/github.com/gvalkov/tailon) [![Go Report Card](https://goreportcard.com/badge/github.com/gvalkov/tailon)](https://goreportcard.com/report/github.com/gvalkov/tailon) [![Apache License](https://img.shields.io/badge/license-Apache-blue.svg)](https://github.com/gvalkov/tailon/blob/master/LICENSE) [![GitHub release](https://img.shields.io/github/release/gvalkov/tailon.svg)](https://github.com/gvalkov/tailon/releases)

Tailon is a webapp for looking at and searching through files and streams. In a
nutshell, it is a fancy web wrapper around the following commands:

```
tail -f
tail -f | grep
tail -f | awk
tail -f | sed
```

What sets tailon apart from other similar projects is:

* Fully self-contained executable.
* Small footprint. The tailon executable sits at 2.5MB in size and uses 10MB of RSS.
* Responsive and minimal user-interface.

## Install

Download a build for your platform from the [releases] page or install using `go get`:

```
go get -u github.com/gvalkov/tailon
```

A docker image is also available:

```
docker run --rm gvalkov/tailon --help
```

## Usage

Tailon is a command-line program that starts a local HTTP server, which in turn
streams the output of commands such as `tail` and `grep`. It can be configured
from its command-line interface or through the convenience of a [toml] config
file. Some options, like adding new commands, are only available through the
configuration file.

To get started, run tailon with the list of files that you wish to monitor.

```
tailon /var/log/apache/access.log /var/log/apache/error.log /var/log/messages
```

Tailon can serve single files, globs or whole directory trees. Tailon’s
server-side functionality is summarized entirely in its help message:

[//]: # (run "make README.md" to update the next section with the output of tailon --help)

[//]: # (BEGIN HELP)
```
Usage: tailon [options] -c <config file>
Usage: tailon [options] <filespec> [<filespec> ...]

Tailon is a webapp for looking at and searching through files and streams.

  -a, --allow-download         allow file downloads (default true)
  -b, --bind string            Listen on the specified address and port (default ":8080")
  -c, --config string
  -h, --help                   Show this help message and exit
  -e, --help-config            Show configuration file help and exit
  -r, --relative-root string   webapp relative root (default "/")

Tailon can be configured through a config file or with command-line flags.

The command-line interface expects one or more filespec arguments, which
specify the files to be served. The expected format is:

  [alias=name,group=name]<spec>

where <spec> can be a file name, glob or directory. The optional 'alias='
and 'group=' specifiers change the display name of the files in the UI and
the group in which they appear.

A file specifier points to a single, possibly non-existent file. The file
name in the UI can be overwritten with 'alias='. For example:

  tailon alias=error.log,/var/log/apache/error.log

A glob evaluates to the list of files that match a shell file name pattern.
The pattern is evaluated each time the file list is refreshed. An 'alias='
specifier overwrites the parent directory of each matched file in the UI.

  tailon "/var/log/apache/*.log" "alias=nginx,/var/log/nginx/*.log"

If a directory is given, all files under it are served recursively.

  tailon /var/log/apache/ /var/log/nginx/

Example usage:
  tailon file1.txt file2.txt file3.txt
  tailon alias=messages,/var/log/messages "/var/log/*.log"
  tailon -b localhost:8080 -c config.toml

For information on usage through the configuration file, please refer to the
'--help-config' option.
```
[//]: # (END HELP)

## Security

Tailon runs commands on the server it is installed on. While commands that
accept a script argument (such as awk, sed and grep) should be invulnerable to
shell injection, they may still allow for arbitrary command execution and
unrestricted access to the filesystem.

To clarify this point, consider the following input to the sed command:

```
s/a/b'; cat /etc/secrets
```

This will result in an error, as tailon does not invoke commands through a
shell. On the other hand, the following command is a perfectly valid sed script
that has the same effect as the above attempt for shell injection:

```
r /etc/secrets
```

The default set of enabled commands - tail, grep and awk - should be safe to
use. GNU awk is run in [sandbox] mode, which prevents scripts from accessing your
system, either through the system() builtin or by using input redirection.

By default, tailon is accessible to anyone who knows the server address and
port. Basic and digest authentication are under development.


## Development

### Frontend

The web interface is a written in plain ES5 with the help of some Vue.js. A
simple makefile is used to bundle and compress the frontend assets. To work on
the frontend, make sure you're building with the `dev` build tag:

```
go build -tags dev
```

This will ensure that the `tailon` binary is reading assets from the
`frontend/dist` directory instead of from `frontend/assets_vfsdata.go`. To
compile the web assets, use `make all` or `make all BUILD=dev` in case you want
to simply concatenate files instead of also compressing them.

The `make watch` goal can be used to continuously update the bundles as you make
changes to the sources.

Note that the minified frontend bundles are committed in order to avoid people
wanting to work only on the backend from having to pull the full `node_modules`.


### Backend

The backend is written in straightforward go that tries to do as much as
possible using only the standard library.


### TODO:

* Directory serving is not implemented yet.

* User specified TOML configuration files.

* Basic and digest authentication.

* Add a 'command' filespec - e.g. `"command,journalctl -u nginx"`.

* Better configuration dialog.

* Add interface themes - e.g. light, dark and solarized.

* Add ability to change font family and size.

* Windows support (can use one of the tail implementations in go)

* Implement [wtee].

* Handling of stderr!


### Testing

The project has unit-tests, which you can run with `go test` and integration
tests which you can run with `cd tests; pytest`. Alternatively, you can run both
with `make test`.

The integration tests are written in Python and use `pytest` and `aiohttp` to
interact with a running `tailon` instance. To run the integration tests, you
first need to install the needed dependencies:

```shell
# Option 1: Using a virtualenv
python3 -m venv path/to/venv
source path/to/venv/bin/activate
python3 -m pip install -r tests/requirements.txt

# Option 2: User site-packages
python3 -m pip install --user -r tests/requirements.txt
```


## What about the other tailon project?

This project is a full rewrite of the original [tailon] with the following goals in mind:

* Easier maintenance for the maintainer.
* Remove unwanted features and fix poor design choices.
* Learn more.

In terms of tech, the following has changed:

* Backend from Python+Tornado to Go.
* Frontend from a very-custom and specific Typescript solution to a simple ES5 + vue.js
  (mostly for data-binding).
* Simplified asset pipeline (a short Makefile).
* Config file is now toml.
* Fully self contained


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
[tailon]:    https://github.com/gvalkov/tailon
[wtee]:      https://github.com/gvalkov/wtee
[toml]:      https://github.com/toml-lang/toml
[releases]:  https://github.com/gvalkov/tailon-next/releases
[errorlog]:  http://www.psychogenic.com/en/products/Errorlog.php
[log.io]:    http://logio.org/
[rtail]:     http://rtail.org/
[this icon]: http://www.iconfinder.com/icondetails/15150/48/terminal_icon
[sandbox]:   http://www.gnu.org/software/gawk/manual/html_node/Options.html#index-g_t_0040code_007b_002dS_007d-option-277
[Apache 2.0 License]: LICENSE
