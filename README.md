Overview
========

properties is a Go library for reading and writing properties files.

It supports reading from multiple files and Spring style recursive property
expansion of expressions like '${key}' to their corresponding value.

The current version supports both ISO-8859-1 and UTF-8 encoded data.

Install
-------

	$ go get github.com/magiconair/properties

Documentation
-------------

See [![GoDoc](https://godoc.org/github.com/magiconair/properties?status.png)](https://godoc.org/github.com/magiconair/properties)

History
-------

v1.1.0, 20 Jan 2014
-------------------
* Renamed from goproperties to properties
* Added support for expansion of environment vars in
  filenames and value expressions

v1.0.0, 7 Jan 2014
------------------
* Initial release

License
-------

2 clause BSD license. See LICENSE file for details.

