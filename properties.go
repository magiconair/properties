// Copyright 2013 Frank Schroeder. All rights reserved. MIT licensed.

package properties

import (
	"fmt"
	"io"
	"io/ioutil"
	"unicode/utf8"
)

type Properties struct {
	m map[string]string
}

// Reads bytes fully and parses them as ISO-8859-1.
func NewProperties(r io.Reader) (*Properties, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return NewPropertiesFromString(toUtf8(buf))
}

func NewPropertiesFromISO8859_1(buf []byte) (*Properties, error) {
	return newParser().Parse(toUtf8(buf))
}

// Java properties spec says that .properties files must be ISO-8859-1
// encoded. Therefore, we first convert them to UTF-8 and then parse them.
func NewPropertiesFromString(input string) (*Properties, error) {
	if err := isISO8859_1(input); err != nil {
		return nil, err
	}
	return newParser().Parse(input)
}

// returns the value for the given key
func (p *Properties) Get(key string) (value string, ok bool) {
	value, ok = p.m[key]
	return value, ok
}

// sets the property key = value and returns the previous value if exists or an empty string
func (p *Properties) Set(key, value string) (prevValue string) {
	prevValue, ok := p.m[key]
	if !ok {
		prevValue = ""
	}

	p.m[key] = value
	return prevValue
}

// returns the number of keys
func (p *Properties) Len() int {
	return len(p.m)
}

// taken from
// http://stackoverflow.com/questions/13510458/golang-convert-iso8859-1-to-utf8
func toUtf8(iso8859_1_buf []byte) string {
	buf := make([]rune, len(iso8859_1_buf))
	for i, b := range iso8859_1_buf {
		buf[i] = rune(b)
	}
	return string(buf)
}

func isISO8859_1(s string) error {
	for i := 0; i < len(s); i++ {
		r, w := utf8.DecodeRuneInString(s[i:])
		if w > 1 || r > 255 {
			return fmt.Errorf("invalid ISO-8859-1 input. %s", s)
		}
	}
	return nil
}
