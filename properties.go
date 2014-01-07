// Copyright 2013 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goproperties

// BUG(frank): Set() does not check for invalid unicode literals since this is currently handled by the lexer.
// BUG(frank): Write() does not allow to configure the newline character. Therefore, on Windows LF is used.

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Properties struct {
	// Pre-/Postfix for property expansion.
	Prefix  string
	Postfix string

	m map[string]string
}

// NewProperties creates a new Properties struct with the default
// configuration for "${key}" expressions.
func NewProperties() *Properties {
	return &Properties{
		Prefix:  "${",
		Postfix: "}",
		m:       make(map[string]string),
	}
}

// Get returns the expanded value for the given key if exists. Otherwise, ok is false.
func (p *Properties) Get(key string) (value string, ok bool) {
	v, ok := p.m[key]
	if !ok {
		return "", false
	}

	expanded, err := p.expand(v)

	// we guarantee that the expanded value is free of
	// circular references and malformed expressions
	// so we panic if we still get an error here.
	if err != nil {
		panic(fmt.Errorf("%s in %q", err, key+" = "+v))
	}

	return expanded, true
}

// GetBool checks if the expanded value is one of '1', 'yes',
// 'true' or 'on' if the key exists. The comparison is case-insensitive.
// If the key does not exist the default value is returned.
func (p *Properties) GetBool(key string, def bool) bool {
	if v, ok := p.Get(key); ok {
		v = strings.ToLower(v)
		return v == "1" || v == "true" || v == "yes" || v == "on"
	}
	return def
}

// GetFloat64 parses the expanded value as a float64 if the key exists.
// If key does not exist or the value cannot be parsed the default
// value is returned.
func (p *Properties) GetFloat64(key string, def float64) float64 {
	if v, ok := p.Get(key); ok {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return def
		}
		return n
	}
	return def
}

// GetInt64 parses the expanded value as an int if the key exists.
// If key does not exist or the value cannot be parsed the default
// value is returned.
func (p *Properties) GetInt64(key string, def int64) int64 {
	if v, ok := p.Get(key); ok {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return def
		}
		return n
	}
	return def
}

// GetUint64 parses the expanded value as an uint64 if the key exists.
// If key does not exist or the value cannot be parsed the default
// value is returned.
func (p *Properties) GetUint64(key string, def uint64) uint64 {
	if v, ok := p.Get(key); ok {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return def
		}
		return n
	}
	return def
}

// GetString returns the expanded value for the given key if exists or the default value otherwise.
func (p *Properties) GetString(key, def string) string {
	if v, ok := p.Get(key); ok {
		return v
	}
	return def
}

// Len returns the number of keys.
func (p *Properties) Len() int {
	return len(p.m)
}

// Set sets the property key to the corresponding value.
// If a value for key existed before then ok is true and prev
// contains the previous value. If the value contains a
// circular reference or a malformed expression then
// an error is returned.
func (p *Properties) Set(key, value string) (prev string, ok bool, err error) {
	_, err = p.expand(value)
	if err != nil {
		return "", false, err
	}

	v, ok := p.Get(key)
	p.m[key] = value
	return v, ok, nil
}

// String returns a string of all expanded 'key = value' pairs.
func (p *Properties) String() string {
	var s string
	for key, _ := range p.m {
		value, _ := p.Get(key)
		s = fmt.Sprintf("%s%s = %s\n", s, key, value)
	}
	return s
}

// Write writes all unexpanded 'key = value' pairs to the given writer.
func (p *Properties) Write(w io.Writer, enc Encoding) (int, error) {
	total := 0
	for key, value := range p.m {
		s := fmt.Sprintf("%s = %s\n", encode(key, " :", enc), encode(value, "", enc))
		n, err := w.Write([]byte(s))
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// ----------------------------------------------------------------------------

// check expands all values and returns an error if a circular reference or
// a malformed expression was found.
func (p *Properties) check() error {
	for _, value := range p.m {
		if _, err := p.expand(value); err != nil {
			return err
		}
	}
	return nil
}

// expand recursively expands expressions of '(prefix)key(postfix)' to their corresponding values.
// The function keeps track of the keys that were already expanded and stops if it
// detects a circular reference or a malformed expression.
func (p *Properties) expand(input string) (string, error) {
	// no pre/postfix -> nothing to expand
	if p.Prefix == "" && p.Postfix == "" {
		return input, nil
	}

	return p.doExpand(input, make(map[string]bool))
}

func (p *Properties) doExpand(s string, keys map[string]bool) (string, error) {
	a := strings.Index(s, p.Prefix)
	if a == -1 {
		return s, nil
	}

	b := strings.Index(s[a:], p.Postfix)
	if b == -1 {
		return "", fmt.Errorf("Malformed expression")
	}

	key := s[a+len(p.Prefix) : b-len(p.Postfix)+1]

	if _, ok := keys[key]; ok {
		return "", fmt.Errorf("Circular reference")
	}

	val, ok := p.m[key]
	if !ok {
		val = ""
	}

	// remember that we've seen the key
	keys[key] = true

	return p.doExpand(s[:a]+val+s[b+1:], keys)
}

// encode encodes a UTF-8 string to ISO-8859-1 and escapes some characters.
func encode(s string, special string, enc Encoding) string {
	switch enc {
	case UTF8:
		return encodeUtf8(s, special)
	case ISO_8859_1:
		return encodeIso(s, special)
	default:
		panic(fmt.Sprintf("Unsupported encoding %v", enc))
	}
}

func encodeUtf8(s string, special string) string {
	v := ""
	for pos := 0; pos < len(s); {
		r, w := utf8.DecodeRuneInString(s[pos:])
		pos += w
		v += escape(r, special)
	}
	return v
}

func encodeIso(s string, special string) string {
	var r rune
	var w int
	var v string
	for pos := 0; pos < len(s); {
		switch r, w = utf8.DecodeRuneInString(s[pos:]); {
		case r < 1<<8: // single byte rune -> escape special chars only
			v += escape(r, special)
		case r < 1<<16: // two byte rune -> unicode literal
			v += fmt.Sprintf("\\u%04x", r)
		default: // more than two bytes per rune -> can't encode
			v += "?"
		}
		pos += w
	}
	return v
}

func escape(r rune, special string) string {
	switch r {
	case '\f':
		return "\\f"
	case '\n':
		return "\\n"
	case '\r':
		return "\\r"
	case '\t':
		return "\\t"
	default:
		if strings.ContainsRune(special, r) {
			return "\\" + string(r)
		}
		return string(r)
	}
}
