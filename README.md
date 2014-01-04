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

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		// handle error
	}

	p, err := goproperties.Decode(buf)
	if err != nil {
		// handle error
	}

	value, ok := p.Get("key")
	if ok {
		fmt.Println(value)
	}

or

	p, err := goproperties.DecodeString("key = value")
	if err != nil {
		// handle error
	}

	value, ok := p.Get("key")
	if ok {
		fmt.Println(value)
	}

History
-------

v0.9, 17 Dec 2013 - Initial release

License
-------

2 clause BSD license. See LICENSE file for details.

Parts of the lexer are taken from the template/text/parser package
For these parts the following applies:

Copyright 2011 The Go Authors. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file of the go 1.2
distribution.
