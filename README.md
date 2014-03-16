Overview [![Build Status](https://travis-ci.org/magiconair/properties.png?branch=master)](https://travis-ci.org/magiconair/properties)
========

properties is a Go library for reading and writing properties files.

It supports reading from multiple files and Spring style recursive property
expansion of expressions like `${key}` to their corresponding value.

Value expressions can refer to other keys like in `${key}` or to
environment variables like in `${USER}`.

Filenames can also contain environment variables like in
`/home/${USER}/myapp.properties`.

The properties library supports both ISO-8859-1 and UTF-8 encoded data.

Install
-------

	$ go get github.com/magiconair/properties

Documentation
-------------

See [![GoDoc](https://godoc.org/github.com/magiconair/properties?status.png)](https://godoc.org/github.com/magiconair/properties)

History
-------

v1.2.0, 05 Mar 2014
-------------------
* Added MustGet... functions
* Added support for int and uint with range checks on 32 bit platforms

v1.1.0, 20 Jan 2014
-------------------
* Renamed from goproperties to properties
* Added support for expansion of environment vars in
  filenames and value expressions
* Fixed bug where value expressions were not at the
  start of the string

v1.0.0, 7 Jan 2014
------------------
* Initial release

License
-------

2 clause BSD license. See LICENSE file for details.

ToDo
----
* Dump contents with passwords and secrets obscured
* panic on non-existent key
* log non-existent key
