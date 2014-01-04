// Copyright 2013 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goproperties

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Load reads an ISO-8859-1 encoded buffer into a Properties struct.
func Load(buf []byte) (*Properties, error) {
	return loadBuf(buf, enc_iso_8859_1)
}

// LoadString reads an UTF-8 string into a Properties struct.
func LoadString(input string) (*Properties, error) {
	return loadBuf([]byte(input), enc_utf8)
}

// LoadFile reads a file into a Properties struct.
func LoadFile(filename string) (*Properties, error) {
	return loadFiles([]string{filename}, false)
}

// LoadFiles reads multiple file in the given order into
// a Properties struct. If 'ignoreMissing' is 'true' then
// non-existent files will not be reported as error.
func LoadFiles(filenames []string, ignoreMissing bool) (*Properties, error) {
	return loadFiles(filenames, ignoreMissing)
}

// MustLoadFile reads a file into a Properties struct and panics on error.
func MustLoadFile(filename string) *Properties {
	return MustLoadFiles([]string{filename}, false)
}

// MustLoadFiles reads multiple file in the given order into
// a Properties struct and panics on error.
// If 'ignoreMissing' is 'true' then non-existent files will not be reported as error.
func MustLoadFiles(filenames []string, ignoreMissing bool) *Properties {
	p, err := loadFiles(filenames, ignoreMissing)
	if err != nil {
		panic(err)
	}
	return p
}

type encoding uint

const (
	enc_utf8 encoding = 1 << iota
	enc_iso_8859_1
)

// Loads either an ISO-8859-1 or an UTF-8 encoded string into a Properties struct.
func loadBuf(buf []byte, enc encoding) (*Properties, error) {
	return parse(convert(buf, enc))
}

func loadFiles(filenames []string, ignoreMissing bool) (*Properties, error) {
	buff := make([]byte, 0, 4096)

	for _, filename := range filenames {
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			if ignoreMissing && os.IsNotExist(err) {
				// TODO(frank): should we log that we are skipping the file?
				continue
			}
			return nil, err
		}

		// concatenate the buffers and add a new line in case
		// the previous file didn't end with a new line
		buff = append(append(buff, buf...), '\n')
	}

	return loadBuf(buff, enc_iso_8859_1)
}

// Interprets a byte buffer either as ISO-8859-1 or UTF-8 encoded string.
// For ISO-8859-1 we can convert each byte straight into a rune since the
// first 256 unicode code points cover ISO-8859-1.
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
