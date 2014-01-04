// Copyright 2013 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goproperties

import (
	"fmt"
	"io"
	"log"
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

	// if there is an error then this is a format exception which we just log
	// and return the input unchanged.
	if err != nil {
		log.Printf("%s in %q", err, key+" = "+v)
		return v, true
	}

	return expanded, true
}

// GetDefault returns the expanded value for the given key if exists or the default value otherwise.
func (p *Properties) GetDefault(key, defaultValue string) (value string) {
	if v, ok := p.Get(key); ok {
		return v
	}
	return defaultValue
}

// Len returns the number of keys.
func (p *Properties) Len() int {
	return len(p.m)
}

// Dump returns a string of all unexpanded 'key = value' pairs.
func (p *Properties) Dump() string {
	var s string
	for key, value := range p.m {
		s = fmt.Sprintf("%s%s = %s\n", s, key, value)
	}
	return s
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

// Write writes all unexpanded 'key = value' pairs as ISO-8859-1 to the given writer.
func (p *Properties) Write(w io.Writer) (int, error) {
	total := 0
	for key, value := range p.m {
		s := fmt.Sprintf("%s = %s\n", encode(key, " :"), encode(value, ""))
		n, err := w.Write([]byte(s))
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// expand recursively expands expressions of '(prefix)key(postfix)' to their corresponding values.
// The function keeps track of the keys that were already expanded and stops if it
// detects a circular reference.
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
func encode(s string, escape string) string {
	var r rune
	var w int
	var v string
	for pos := 0; pos < len(s); {
		switch r, w = utf8.DecodeRuneInString(s[pos:]); {
		case r < 1<<8: // single byte rune -> encode special chars only
			switch r {
			case '\f':
				v += "\\f"
			case '\n':
				v += "\\n"
			case '\r':
				v += "\\r"
			case '\t':
				v += "\\t"
			default:
				if strings.ContainsRune(escape, r) {
					v += "\\"
				}
				v += string(r)
			}
		case r < 1<<16: // two byte rune -> unicode literal
			v += fmt.Sprintf("\\u%04x", r)
		default: // more than two bytes per rune -> can't encode
			v += "?"
		}
		pos += w
	}
	return v
}
