Overview
========

goproperties is a Go library for reading and writing Java property files.

It supports reading properties from multiple files and Spring style property
expansion of expressions of '${key}' to their corresponding value.

The current version supports reading both ISO-8859-1 and UTF-8 encoded data.

Install
-------

	$ go get github.com/magiconair/goproperties

Usage
-----

	p := goproperties.MustLoadFile(filename, goproperties.ISO_8859_1)
	value, ok := p.Get("key")
	if ok {
		fmt.Println(value)
	}

or

	// load multiple files and ignore missing files
	p := goproperties.MustLoadFiles([]string{filename1, filename2}, true, goproperties.ISO_8859_1)
	value, ok := p.Get("key")
	if ok {
		fmt.Println(value)
	}

History
-------

v1.0, 6 Jan 2014 - Initial release

License
-------

2 clause BSD license. See LICENSE file for details.

