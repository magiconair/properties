// Copyright 2013 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package goproperties reads Java properties files.
//
// Java properties files contain key/value pairs in one of the following form:
//
//   key value
//   key = value
//   key : value
//
// The value is optional and ends with EOF or a new line which can either be '\n', '\r' or "\r\n".
// Therefore, the following expression is legal and results in a key with an empty value:
//
//   key
//
// Whitespace before the key and around the delimiter is ignored. Whitespace at the end of the value is part of the value.
// Besides the space ' ' (U+0020) character the TAB (U+0009) and FF (U+000C) characters are also treated as whitespace.
// Therefore, the following expressions are equal:
//
//   key=value
//      key=value
//   key= value
//   key =value
//   key = value
//   key   =   value
//   key\f=\fvalue
//   key\t=\tvalue
//
// Blank lines and comment lines starting with '#' or '!' and are ignored until the end of the line.
//
//   # the next line is empty and will be ignored
//
//   ! this is a comment
//   key = value
//
// If the delimiter characters '=' and ':' appear in either key or value then
// they must be escaped with a backslash. Because of this the backslash must
// also be escaped. The characters '\n', '\r' or '\t' can be part of both key
// or value and must be escaped. For all other characters the backslash is
// silently dropped.
//
//   # key:1 = value=2
//   key\:1 = value\=2
//
//   # key = value	with	tabs
//   key = value\twith\ttabs
//
//   # key = value with silently dropped backslash
//   key = v\alu\e with silently dropped backslash
//
// Values can span multiple lines by using a backslash before the newline character.
// All subsequent whitespace on the following line is ignored. Comment lines cannot be
// extended like this.
//
//   # key = value continued
//   key = value \
//         continued
//
// Java properties files are ISO-8559-1 encoded and can have Unicode literals for
// characters outside the character set. Unicode literals are specified as \uXXXX.
//
//   # key = value with €
//   key = value with \u20AC
//
package goproperties
