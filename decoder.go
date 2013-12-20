// Copyright 2013 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goproperties

import (
	"fmt"
)

type encoding uint

const (
	enc_utf8 encoding = 1 << iota
	enc_iso_8859_1
)

// Decodes an ISO-8859-1 encoded buffer into a Properties struct.
func Decode(buf []byte) (Properties, error) {
	return decodeWithEncoding(buf, enc_iso_8859_1)
}

// Decodes an UTF-8 string into a Properties struct.
func DecodeFromString(input string) (Properties, error) {
	return decodeWithEncoding([]byte(input), enc_utf8)
}

// Decodes either an ISO-8859-1 or an UTF-8 encoded string into a Properties struct.
func decodeWithEncoding(buf []byte, enc encoding) (Properties, error) {
	return newParser().Parse(convert(buf, enc))
}

// The Java properties spec says that .properties files must be ISO-8859-1
// encoded. Since the first 256 unicode code points cover ISO-8859-1 we
// can convert each byte straight into a rune and use the resulting string
// as UTF-8 input for the parser.
func convert(buf []byte, enc encoding) string {
	switch enc {
	case enc_utf8:
		return string(buf)
	case enc_iso_8859_1:
		runes := make([]rune, len(buf))
		for i, b := range buf {
			runes[i] = rune(b)
		}
		return string(runes)
	default:
		panic(fmt.Sprintf("unsupported encoding %v", enc))
	}
}
