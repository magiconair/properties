// Copyright 2013 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// goproperties provides functions for reading and writing
// ISO-8859-1 and UTF-8 encoded .properties files and has
// support for recursive property expansion.
//
// Java properties files are ISO-8859-1 encoded and use Unicode
// literals for characters outside the ISO character set. Unicode
// literals can be used in UTF-8 encoded properties files but
// aren't necessary.
//
// All of the different key/value delimiters ' ', ':' and '=' are
// supported as well as the comment characters '!' and '#' and
// multi-line values.
//
//   ! this is a comment
//   # and so is this
//
//   # the following expressions are equal
//   key value
//   key=value
//   key:value
//   key = value
//   key : value
//   key = val\
//         ue
//
// Property expansion is recursive and circular references
// and malformed expressions are not allowed and cause an
// error.
//
//   # standard property
//   key = value
//
//   # property expansion: key2 = value
//   key2 = ${key}
//
//   # recursive expansion: key3 = value
//   key3 = ${key2}
//
//   # circular reference (error)
//   key = ${key}
//
//   # malformed expression (error)
//   key = ${ke
//
// The default property expansion format is ${key} but can be
// changed by setting different pre- and postfix values on the
// Properties object.
//
//   p := goproperties.NewProperties()
//   p.Prefix = "#["
//   p.Postfix = "]#"
//
// Properties provides convenience functions for getting typed
// values with default values if the key does not exist or the
// type conversion failed.
//
//   # Returns true if the value is either "1", "on", "yes" or "true"
//   # Returns false for every other value and the default value if
//   # the key does not exist.
//   v = p.GetBool("key", false)
//
//   # Returns the value if the key exists and the format conversion
//   # was successful. Otherwise, the default value is returned.
//   v = p.GetInt64("key", 999)
//   v = p.GetUint64("key", 999)
//   v = p.GetFloat64("key", 123.0)
//   v = p.GetString("key", "def")
//
// The following documents provide a description of the properties
// file format.
//
// http://en.wikipedia.org/wiki/.properties
//
// http://docs.oracle.com/javase/7/docs/api/java/util/Properties.html#load%28java.io.Reader%29
//
package goproperties
