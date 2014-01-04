// Copyright 2013 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goproperties

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	. "launchpad.net/gocheck"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var (
	_       = Suite(&TestSuite{})
	verbose = flag.Bool("verbose", false, "Verbose output")
)

// define test cases in the form of
// {"input", "key1", "value1", "key2", "value2", ...}
var complexTests = [][]string{
	// whitespace prefix
	{" key=value", "key", "value"},     // SPACE prefix
	{"\fkey=value", "key", "value"},    // FF prefix
	{"\tkey=value", "key", "value"},    // TAB prefix
	{" \f\tkey=value", "key", "value"}, // mix prefix

	// multiple keys
	{"key1=value1\nkey2=value2\n", "key1", "value1", "key2", "value2"},
	{"key1=value1\rkey2=value2\r", "key1", "value1", "key2", "value2"},
	{"key1=value1\r\nkey2=value2\r\n", "key1", "value1", "key2", "value2"},

	// blank lines
	{"\nkey=value\n", "key", "value"},
	{"\rkey=value\r", "key", "value"},
	{"\r\nkey=value\r\n", "key", "value"},

	// escaped chars in key
	{"k\\ ey = value", "k ey", "value"},
	{"k\\:ey = value", "k:ey", "value"},
	{"k\\=ey = value", "k=ey", "value"},
	{"k\\fey = value", "k\fey", "value"},
	{"k\\ney = value", "k\ney", "value"},
	{"k\\rey = value", "k\rey", "value"},
	{"k\\tey = value", "k\tey", "value"},

	// escaped chars in value
	{"key = v\\ alue", "key", "v alue"},
	{"key = v\\:alue", "key", "v:alue"},
	{"key = v\\=alue", "key", "v=alue"},
	{"key = v\\falue", "key", "v\falue"},
	{"key = v\\nalue", "key", "v\nalue"},
	{"key = v\\ralue", "key", "v\ralue"},
	{"key = v\\talue", "key", "v\talue"},

	// silently dropped escape character
	{"k\\zey = value", "kzey", "value"},
	{"key = v\\zalue", "key", "vzalue"},

	// unicode literals
	{"key\\u2318 = value", "key⌘", "value"},
	{"k\\u2318ey = value", "k⌘ey", "value"},
	{"key = value\\u2318", "key", "value⌘"},
	{"key = valu\\u2318e", "key", "valu⌘e"},

	// multiline values
	{"key = valueA,\\\n    valueB", "key", "valueA,valueB"},   // SPACE indent
	{"key = valueA,\\\n\f\f\fvalueB", "key", "valueA,valueB"}, // FF indent
	{"key = valueA,\\\n\t\t\tvalueB", "key", "valueA,valueB"}, // TAB indent
	{"key = valueA,\\\n \f\tvalueB", "key", "valueA,valueB"},  // mix indent

	// comments
	{"# this is a comment\n! and so is this\nkey1=value1\nkey#2=value#2\n\nkey!3=value!3\n# and another one\n! and the final one", "key1", "value1", "key#2", "value#2", "key!3", "value!3"},

	// expansion tests
	{"key=value\nkey2=${key}", "key", "value", "key2", "value"},
	{"key=value\nkey2=${key}\nkey3=${key2}", "key", "value", "key2", "value", "key3", "value"},

	// circular references
	{"key=${key}", "key", "${key}"},
	{"key1=${key2}\nkey2=${key1}", "key1", "${key2}", "key2", "${key1}"},

	// malformed expressions
	{"key=${ke", "key", "${ke"},
	{"key=valu${ke", "key", "valu${ke"},
}

// define error test cases in the form of
// {"input", "expected error message"}
var errorTests = [][]string{
	{"key\\u1 = value", "invalid unicode literal"},
	{"key\\u12 = value", "invalid unicode literal"},
	{"key\\u123 = value", "invalid unicode literal"},
	{"key\\u123g = value", "invalid unicode literal"},
	{"key\\u123", "invalid unicode literal"},
}

// define write encoding test cases in the form of
// {"input", "expected output after write"}
var writeTests = [][]string{
	{"key = value", "key = value\n"},
	{"key = value \\\n   continued", "key = value continued\n"},
	{"key⌘ = value", "key\\u2318 = value\n"},
	{"ke\\ \\:y = value", "ke\\ \\:y = value\n"},
}

// Benchmarks the decoder by creating a property file with 1000 key/value pairs.
func BenchmarkDecoder(b *testing.B) {
	input := ""
	for i := 0; i < 1000; i++ {
		input += fmt.Sprintf("key%d=value%d\n", i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load([]byte(input), ISO_8859_1)
	}
}

// tests basic single key/value combinations with all possible whitespace, delimiter and newline permutations.
func (l *TestSuite) TestBasic(c *C) {
	testAllCombinations(c, "key", "")
	testAllCombinations(c, "key", "value")
	testAllCombinations(c, "key", "value   ")
}

// tests more complex cases.
func (l *TestSuite) TestComplex(c *C) {
	for _, test := range complexTests {
		testKeyValue(c, test[0], test[1:]...)
	}
}

// tests error cases.
func (l *TestSuite) TestErrors(c *C) {
	for _, test := range errorTests {
		input, msg := test[0], test[1]
		testError(c, input, msg)
	}
}

// Test write encoding.
func (l *TestSuite) TestWrite(c *C) {
	for _, test := range writeTests {
		input, output := test[0], test[1]
		p, err := parse(input)

		buf := new(bytes.Buffer)
		n, err := p.Write(buf)
		c.Assert(err, IsNil)
		s := string(buf.Bytes())
		c.Assert(n, Equals, len(output), Commentf("input=%q expected=%q obtained=%q", input, output, s))
		c.Assert(s, Equals, output, Commentf("input=%q expected=%q obtained=%q", input, output, s))
	}
}

// tests all combinations of delimiters, leading and/or trailing whitespace and newlines.
func testAllCombinations(c *C, key, value string) {
	whitespace := []string{"", " ", "\f", "\t"}
	delimiters := []string{"", " ", "=", ":"}
	newlines := []string{"", "\r", "\n", "\r\n"}
	for _, dl := range delimiters {
		for _, ws1 := range whitespace {
			for _, ws2 := range whitespace {
				for _, nl := range newlines {
					// skip the one case where there is nothing between a key and a value
					if ws1 == "" && dl == "" && ws2 == "" && value != "" {
						continue
					}

					input := fmt.Sprintf("%s%s%s%s%s%s", key, ws1, dl, ws2, value, nl)
					testKeyValue(c, input, key, value)
				}
			}
		}
	}
}

// tests whether key/value pairs exist for a given input.
// keyvalues is expected to be an even number of strings of "key", "value", ...
func testKeyValue(c *C, input string, keyvalues ...string) {
	printf("%q\n", input)

	p, err := Load([]byte(input), ISO_8859_1)
	c.Assert(err, IsNil)
	assertKeyValues(c, input, p, keyvalues...)
}

// tests whether some input produces a given error message.
func testError(c *C, input, msg string) {
	printf("%q\n", input)

	_, err := Load([]byte(input), ISO_8859_1)
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), msg), Equals, true, Commentf("Expected %q got %q", msg, err.Error()))
}

// tests whether key/value pairs exist for a given input.
// keyvalues is expected to be an even number of strings of "key", "value", ...
func assertKeyValues(c *C, input string, p *Properties, keyvalues ...string) {
	c.Assert(p, NotNil)
	c.Assert(2*p.Len(), Equals, len(keyvalues), Commentf("Odd number of key/value pairs."))

	for i := 0; i < len(keyvalues); i += 2 {
		key, value := keyvalues[i], keyvalues[i+1]
		v, ok := p.Get(key)
		c.Assert(ok, Equals, true, Commentf("No key %q found (input=%q)", key, input))
		c.Assert(v, Equals, value, Commentf("Value %q does not match %q (input=%q)", v, value, input))
	}
}

// prints to stderr if the -verbose flag was given.
func printf(format string, args ...interface{}) {
	if *verbose {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}
