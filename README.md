Overview
========

goproperties is a Go library for parsing Java property files.

The current version supports reading both ISO-8859-1 and UTF-8 encoded data.

A future version will also support Spring Framework style property expansion like

	key = value
	key2 = ${key}

History
=======

v0.9 - Initial release

Usage
=====

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		// handle error
	}

	p, err := goproperties.Decode(buf)
	if err != nil {
		// handle error
	}

	value, ok := p.Get("key")

Import
======

	go get github.com/magiconair/goproperties


