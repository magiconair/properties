// Copyright 2013 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// goproperties provides functions for reading
// ISO-8859-1 (Java) and UTF-8 encoded .properties files and has
// support for Spring-like property expansion.
//
//   key = value
//   # key2 = value
//   key2 = ${key}
//
// The default property expansion format is ${key} but can be
// changed by setting different pre- and postfix values on the
// Properties object.
//
//   p := goproperties.NewProperties()
//   p.Prefix = "#["
//   p.Postfix = "]#"
//
// Property expansion is recursive and circular references are not allowed.
// If a circular reference is detected an error is logged and the
// unexpanded value is returned.
//
//   # Circular reference
//   key = ${key}
//
//   # Malformed expression
//   key = ${ke
//
// When writing properties to a writer currently only ISO-8859-1 encoding
// is supported.
//
// See one of the following links for a description of the properties
// file format.
//
// http://en.wikipedia.org/wiki/.properties
//
// http://docs.oracle.com/javase/7/docs/api/java/util/Properties.html#load%28java.io.Reader%29
//
package goproperties
