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

Starting from version 1.3.0 the behavior of the MustXXX() functions is
configurable by providing a custom `ErrorHandler` function. The default has
changed from `panic` to `log.Fatal` but this is configurable and custom
error handling functions can be provided. See the package documentation for
details.

Getting Started
---------------

```go
import "github.com/magiconair/properties"

func main() {
	p := properties.MustLoadFile("${HOME}/config.properties", properties.UTF8)
	host := p.MustGetString("host")
	port := p.GetInt("port", 8080)
}

```

Read the full documentation on [GoDoc](https://godoc.org/github.com/magiconair/properties)   [![GoDoc](https://godoc.org/github.com/magiconair/properties?status.png)](https://godoc.org/github.com/magiconair/properties)

Installation and Upgrade
------------------------

```
$ go get -u github.com/magiconair/properties
```

History
-------

v1.4.0, 23 Sep 2014
-------------------
 * Added Keys() to get the keys
 * Added Filter(), FilterRegexp() and FilterPrefix() to get a subset of the properties

v1.3.0, 18 Mar 2014
-------------------
* Added support for time.Duration
* Made MustXXX() failure behavior configurable (log.Fatal, panic, custom)
* Changed default of MustXXX() failure from panic to log.Fatal

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
