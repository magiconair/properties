// Copyright 2013 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// goproperties provides functions for reading and writing
// Java properties files and has support for Spring like
// property expansion.
//
// By default, if a value contains a reference '${key}' then
// getting the value will recursively expand the key to its value.
// The format is configurable and circular references are not allowed.
//
// See one of the following links for a description of the properties
// file format.
//
// http://en.wikipedia.org/wiki/.properties
//
// http://docs.oracle.com/javase/7/docs/api/java/util/Properties.html#load%28java.io.Reader%29
//
package goproperties
