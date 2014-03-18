// Copyright 2013 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package properties

// BUG(frank): Set() does not check for invalid unicode literals since this is currently handled by the lexer.
// BUG(frank): Write() does not allow to configure the newline character. Therefore, on Windows LF is used.

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
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

// Get returns the expanded value for the given key if exists.
// Otherwise, ok is false.
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

// MustGet returns the expanded value for the given key if exists.
// Otherwise, it panics.
func (p *Properties) MustGet(key string) string {
	if v, ok := p.Get(key); ok {
		return v
	}
	panic(invalidKeyError(key))
}

// ----------------------------------------------------------------------------

// GetBool checks if the expanded value is one of '1', 'yes',
// 'true' or 'on' if the key exists. The comparison is case-insensitive.
// If the key does not exist the default value is returned.
func (p *Properties) GetBool(key string, def bool) bool {
	v, err := p.getBool(key)
	if err != nil {
		return def
	}
	return v
}

// MustGetBool checks if the expanded value is one of '1', 'yes',
// 'true' or 'on' if the key exists. The comparison is case-insensitive.
// If the key does not exist the function panics.
func (p *Properties) MustGetBool(key string) bool {
	v, err := p.getBool(key)
	if err != nil {
		panic(err)
	}
	return v
}

func (p *Properties) getBool(key string) (value bool, err error) {
	if v, ok := p.Get(key); ok {
		v = strings.ToLower(v)
		return v == "1" || v == "true" || v == "yes" || v == "on", nil
	}
	return false, invalidKeyError(key)
}

// ----------------------------------------------------------------------------

// GetDuration parses the expanded value as an time.Duration if the key exists.
// If key does not exist or the value cannot be parsed the default
// value is returned.
func (p *Properties) GetDuration(key string, def time.Duration) time.Duration {
	v, err := p.getInt64(key)
	if err != nil {
		return def
	}
	return time.Duration(v)
}

// MustGetDuration parses the expanded value as an time.Duration if the key exists.
// If key does not exist or the value cannot be parsed the function panics.
func (p *Properties) MustGetDuration(key string) time.Duration {
	v, err := p.getInt64(key)
	if err != nil {
		panic(err)
	}
	return time.Duration(v)
}

// ----------------------------------------------------------------------------

// GetFloat64 parses the expanded value as a float64 if the key exists.
// If key does not exist or the value cannot be parsed the default
// value is returned.
func (p *Properties) GetFloat64(key string, def float64) float64 {
	v, err := p.getFloat64(key)
	if err != nil {
		return def
	}
	return v
}

// GetFloat64 parses the expanded value as a float64 if the key exists.
// If key does not exist or the value cannot be parsed the function panics.
func (p *Properties) MustGetFloat64(key string) float64 {
	v, err := p.getFloat64(key)
	if err != nil {
		panic(err)
	}
	return v
}

func (p *Properties) getFloat64(key string) (value float64, err error) {
	if v, ok := p.Get(key); ok {
		value, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, err
		}
		return value, nil
	}
	return 0, invalidKeyError(key)
}

// ----------------------------------------------------------------------------

// GetInt parses the expanded value as an int if the key exists.
// If key does not exist or the value cannot be parsed the default
// value is returned. If the value does not fit into an int the
// function panics with an out of range error.
func (p *Properties) GetInt(key string, def int) int {
	v, err := p.getInt64(key)
	if err != nil {
		return def
	}
	return intRangeCheck(key, v)
}

// MustGetInt parses the expanded value as an int if the key exists.
// If key does not exist or the value cannot be parsed the function panics.
// If the value does not fit into an int the function panics with
// an out of range error.
func (p *Properties) MustGetInt(key string) int {
	v, err := p.getInt64(key)
	if err != nil {
		panic(err)
	}
	return intRangeCheck(key, v)
}

// ----------------------------------------------------------------------------

// GetInt64 parses the expanded value as an int64 if the key exists.
// If key does not exist or the value cannot be parsed the default
// value is returned.
func (p *Properties) GetInt64(key string, def int64) int64 {
	v, err := p.getInt64(key)
	if err != nil {
		return def
	}
	return v
}

// MustGetInt64 parses the expanded value as an int if the key exists.
// If key does not exist or the value cannot be parsed the function panics.
func (p *Properties) MustGetInt64(key string) int64 {
	v, err := p.getInt64(key)
	if err != nil {
		panic(err)
	}
	return v
}

func (p *Properties) getInt64(key string) (value int64, err error) {
	if v, ok := p.Get(key); ok {
		value, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return value, nil
	}
	return 0, invalidKeyError(key)
}

// ----------------------------------------------------------------------------

// GetUint parses the expanded value as an uint if the key exists.
// If key does not exist or the value cannot be parsed the default
// value is returned. If the value does not fit into an int the
// function panics with an out of range error.
func (p *Properties) GetUint(key string, def uint) uint {
	v, err := p.getUint64(key)
	if err != nil {
		return def
	}
	return uintRangeCheck(key, v)
}

// MustGetUint parses the expanded value as an int if the key exists.
// If key does not exist or the value cannot be parsed the function panics.
// If the value does not fit into an int the function panics with
// an out of range error.
func (p *Properties) MustGetUint(key string) uint {
	v, err := p.getUint64(key)
	if err != nil {
		panic(err)
	}
	return uintRangeCheck(key, v)
}

// ----------------------------------------------------------------------------

// GetUint64 parses the expanded value as an uint64 if the key exists.
// If key does not exist or the value cannot be parsed the default
// value is returned.
func (p *Properties) GetUint64(key string, def uint64) uint64 {
	v, err := p.getUint64(key)
	if err != nil {
		return def
	}
	return v
}

// MustGetUint64 parses the expanded value as an int if the key exists.
// If key does not exist or the value cannot be parsed the function panics.
func (p *Properties) MustGetUint64(key string) uint64 {
	v, err := p.getUint64(key)
	if err != nil {
		panic(err)
	}
	return v
}

func (p *Properties) getUint64(key string) (value uint64, err error) {
	if v, ok := p.Get(key); ok {
		value, err = strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return value, nil
	}
	return 0, invalidKeyError(key)
}

// ----------------------------------------------------------------------------

// GetString returns the expanded value for the given key if exists or
// the default value otherwise.
func (p *Properties) GetString(key, def string) string {
	if v, ok := p.Get(key); ok {
		return v
	}
	return def
}

// MustGetString returns the expanded value for the given key if exists or
// panics otherwise.
func (p *Properties) MustGetString(key string) string {
	if v, ok := p.Get(key); ok {
		return v
	}
	panic(invalidKeyError(key))
}

// ----------------------------------------------------------------------------

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

func (p *Properties) expand(input string) (string, error) {
	// no pre/postfix -> nothing to expand
	if p.Prefix == "" && p.Postfix == "" {
		return input, nil
	}

	return expand(input, make(map[string]bool), p.Prefix, p.Postfix, p.m)
}

// expand recursively expands expressions of '(prefix)key(postfix)' to their corresponding values.
// The function keeps track of the keys that were already expanded and stops if it
// detects a circular reference or a malformed expression of the form '(prefix)key'.
func expand(s string, keys map[string]bool, prefix, postfix string, values map[string]string) (string, error) {
	start := strings.Index(s, prefix)
	if start == -1 {
		return s, nil
	}

	keyStart := start + len(prefix)
	keyLen := strings.Index(s[keyStart:], postfix)
	if keyLen == -1 {
		return "", fmt.Errorf("Malformed expression")
	}

	end := keyStart + keyLen + len(postfix) - 1
	key := s[keyStart : keyStart+keyLen]

	// fmt.Printf("s:%q pp:%q start:%d end:%d keyStart:%d keyLen:%d key:%q\n", s, prefix + "..." + postfix, start, end, keyStart, keyLen, key)

	if _, ok := keys[key]; ok {
		return "", fmt.Errorf("Circular reference")
	}

	val, ok := values[key]
	if !ok {
		val = os.Getenv(key)
	}

	// remember that we've seen the key
	keys[key] = true

	return expand(s[:start]+val+s[end+1:], keys, prefix, postfix, values)
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

func invalidKeyError(key string) error {
	return fmt.Errorf("invalid key: %s", key)
}
