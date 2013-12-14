// Copyright 2013 Frank Schroeder. All rights reserved. MIT licensed.

// Package properties reads Java properties files.
//
// Java properties files contain key/value pairs in one of the following form:
//
//   key = value
//   key : value
//
// Whitespace around the delimiter is ignored which means that the following expressions are equal
//
//   key=value
//   key= value
//   key =value
//   key = value
//   key   =   value
//
// Blank lines and lines starting with '#' or '!' and are ignored until the end of the line.
//
//   # the next line is empty and will be ignored
//
//   ! this is a comment
//   key = value
//
// If the delimiter characters '=' and ':' appear in either key or value then
// they must be escaped with a backslash. Because of this the backslash must
// also be escaped. The characters '\n', '\r' or '\t' can be included in both
// key and value and will be replaced with their correpsonding character.
//
//   # key:1 = value=2
//   key\:1 = value\=2
//
//   # key = value	with	tabs
//   key = value\twith\ttabs
//
// Values can span multiple lines by using a backslash before the newline character.
// All subsequent whitespace on the following line is ignored.
//
//   # key = value continued
//   key = value \
//         continued
//
// Java properties files must be ISO-8559-1 encoded and can have Unicode literals for
// characters outside the character set. Both \uXXXX and \UXXXX are accepted.
//
//   # key = value with €
//   key = value with \U20AC
//
package properties
