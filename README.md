# Tailon-next

The next iteration of the [tailon] log file viewer and [wtee] pipe viewer.
__Currently under slow but steady development__.

## What's new?

Tailon-next is a full rewrite of the original tailon with the following goals in mind:

* Make maintenance easier for me.
* Remove unwanted features and fix poor design choices.
* Learn more stuff.

In terms of tech, the following has changed:

* Backend from Python+Tornado to Go.
* Frontend from a very-custom and specific Typescript solution to a simple ES5 + vue.js
  (mostly for data-binding).
* Simplified asset pipeline (a short Makefile).
* Config file is now toml.

## License

Tailon is released under the terms of the [Apache 2.0 License].

[tailon]: https://github.com/gvalkov/tailon
[wtee]:   https://github.com/gvalkov/wtee
[Apache 2.0 License]: LICENSE
