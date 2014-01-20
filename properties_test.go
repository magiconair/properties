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

// ----------------------------------------------------------------------------

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
	{"key=${USER}", "key", os.Getenv("USER")},
}

// define error test cases in the form of
// {"input", "expected error message"}
var errorTests = [][]string{
	// unicode literals
	{"key\\u1 = value", "Invalid unicode literal"},
	{"key\\u12 = value", "Invalid unicode literal"},
	{"key\\u123 = value", "Invalid unicode literal"},
	{"key\\u123g = value", "Invalid unicode literal"},
	{"key\\u123", "Invalid unicode literal"},

	// circular references
	{"key=${key}", "Circular reference"},
	{"key1=${key2}\nkey2=${key1}", "Circular reference"},

	// malformed expressions
	{"key=${ke", "Malformed expression"},
	{"key=valu${ke", "Malformed expression"},
}

// define write encoding test cases in the form of
// {"input", "expected output after write", ["UTF-8", "ISO-8859-1"]}
var writeTests = [][]string{
	// ISO-8859-1 tests
	{"key = value", "key = value\n", "ISO-8859-1"},
	{"key = value \\\n   continued", "key = value continued\n", "ISO-8859-1"},
	{"key⌘ = value", "key\\u2318 = value\n", "ISO-8859-1"},
	{"ke\\ \\:y = value", "ke\\ \\:y = value\n", "ISO-8859-1"},

	// UTF-8 tests
	{"key = value", "key = value\n", "UTF-8"},
	{"key = value \\\n   continued", "key = value continued\n", "UTF-8"},
	{"key⌘ = value⌘", "key⌘ = value⌘\n", "UTF-8"},
	{"ke\\ \\:y = value", "ke\\ \\:y = value\n", "UTF-8"},
}

// ----------------------------------------------------------------------------

type boolTest struct {
	input, key string
	def, value bool
}

var boolTests = []*boolTest{
	// valid values for TRUE
	&boolTest{"key = 1", "key", false, true},
	&boolTest{"key = on", "key", false, true},
	&boolTest{"key = On", "key", false, true},
	&boolTest{"key = ON", "key", false, true},
	&boolTest{"key = true", "key", false, true},
	&boolTest{"key = True", "key", false, true},
	&boolTest{"key = TRUE", "key", false, true},
	&boolTest{"key = yes", "key", false, true},
	&boolTest{"key = Yes", "key", false, true},
	&boolTest{"key = YES", "key", false, true},

	// valid values for FALSE (all other)
	&boolTest{"key = 0", "key", true, false},
	&boolTest{"key = off", "key", true, false},
	&boolTest{"key = false", "key", true, false},
	&boolTest{"key = no", "key", true, false},

	// non existent key
	&boolTest{"key = true", "key2", false, false},
}

// ----------------------------------------------------------------------------

type floatTest struct {
	input, key string
	def, value float64
}

var floatTests = []*floatTest{
	// valid values
	&floatTest{"key = 1.0", "key", 999, 1.0},
	&floatTest{"key = 0.0", "key", 999, 0.0},
	&floatTest{"key = -1.0", "key", 999, -1.0},
	&floatTest{"key = 1", "key", 999, 1},
	&floatTest{"key = 0", "key", 999, 0},
	&floatTest{"key = -1", "key", 999, -1},
	&floatTest{"key = 0123", "key", 999, 123},

	// invalid values
	&floatTest{"key = 0xff", "key", 999, 999},
	&floatTest{"key = a", "key", 999, 999},

	// non existent key
	&floatTest{"key = 1", "key2", 999, 999},
}

// ----------------------------------------------------------------------------

type intTest struct {
	input, key string
	def, value int64
}

var intTests = []*intTest{
	// valid values
	&intTest{"key = 1", "key", 999, 1},
	&intTest{"key = 0", "key", 999, 0},
	&intTest{"key = -1", "key", 999, -1},
	&intTest{"key = 0123", "key", 999, 123},

	// invalid values
	&intTest{"key = 0xff", "key", 999, 999},
	&intTest{"key = 1.0", "key", 999, 999},
	&intTest{"key = a", "key", 999, 999},

	// non existent key
	&intTest{"key = 1", "key2", 999, 999},
}

// ----------------------------------------------------------------------------

type uintTest struct {
	input, key string
	def, value uint64
}

var uintTests = []*uintTest{
	// valid values
	&uintTest{"key = 1", "key", 999, 1},
	&uintTest{"key = 0", "key", 999, 0},
	&uintTest{"key = 0123", "key", 999, 123},

	// invalid values
	&uintTest{"key = -1", "key", 999, 999},
	&uintTest{"key = 0xff", "key", 999, 999},
	&uintTest{"key = 1.0", "key", 999, 999},
	&uintTest{"key = a", "key", 999, 999},

	// non existent key
	&uintTest{"key = 1", "key2", 999, 999},
}

// ----------------------------------------------------------------------------

type stringTest struct {
	input, key string
	def, value string
}

var stringTests = []*stringTest{
	// valid values
	&stringTest{"key = abc", "key", "def", "abc"},

	// non existent key
	&stringTest{"key = abc", "key2", "def", "def"},
}

// ----------------------------------------------------------------------------

// TestBasic tests basic single key/value combinations with all possible
// whitespace, delimiter and newline permutations.
func (l *TestSuite) TestBasic(c *C) {
	testWhitespaceAndDelimiterCombinations(c, "key", "")
	testWhitespaceAndDelimiterCombinations(c, "key", "value")
	testWhitespaceAndDelimiterCombinations(c, "key", "value   ")
}

func (l *TestSuite) TestComplex(c *C) {
	for _, test := range complexTests {
		testKeyValue(c, test[0], test[1:]...)
	}
}

func (l *TestSuite) TestErrors(c *C) {
	for _, test := range errorTests {
		input, msg := test[0], test[1]
		testError(c, input, msg)
	}
}

func (l *TestSuite) TestGetBool(c *C) {
	for _, test := range boolTests {
		p, err := parse(test.input)
		c.Assert(err, IsNil)
		c.Assert(p.Len(), Equals, 1)
		c.Assert(p.GetBool(test.key, test.def), Equals, test.value)
	}
}

func (l *TestSuite) TestGetFloat64(c *C) {
	for _, test := range floatTests {
		p, err := parse(test.input)
		c.Assert(err, IsNil)
		c.Assert(p.Len(), Equals, 1)
		c.Assert(p.GetFloat64(test.key, test.def), Equals, test.value)
	}
}

func (l *TestSuite) TestGetInt64(c *C) {
	for _, test := range intTests {
		p, err := parse(test.input)
		c.Assert(err, IsNil)
		c.Assert(p.Len(), Equals, 1)
		c.Assert(p.GetInt64(test.key, test.def), Equals, test.value)
	}
}

func (l *TestSuite) TestGetUint64(c *C) {
	for _, test := range uintTests {
		p, err := parse(test.input)
		c.Assert(err, IsNil)
		c.Assert(p.Len(), Equals, 1)
		c.Assert(p.GetUint64(test.key, test.def), Equals, test.value)
	}
}

func (l *TestSuite) TestGetString(c *C) {
	for _, test := range stringTests {
		p, err := parse(test.input)
		c.Assert(err, IsNil)
		c.Assert(p.Len(), Equals, 1)
		c.Assert(p.GetString(test.key, test.def), Equals, test.value)
	}
}

func (l *TestSuite) TestWrite(c *C) {
	for _, test := range writeTests {
		input, output, enc := test[0], test[1], test[2]
		p, err := parse(input)

		buf := new(bytes.Buffer)
		var n int
		switch enc {
		case "UTF-8":
			n, err = p.Write(buf, UTF8)
		case "ISO-8859-1":
			n, err = p.Write(buf, ISO_8859_1)
		}
		c.Assert(err, IsNil)
		s := string(buf.Bytes())
		c.Assert(n, Equals, len(output), Commentf("input=%q expected=%q obtained=%q", input, output, s))
		c.Assert(s, Equals, output, Commentf("input=%q expected=%q obtained=%q", input, output, s))
	}
}

// ----------------------------------------------------------------------------

// tests all combinations of delimiters, leading and/or trailing whitespace and newlines.
func testWhitespaceAndDelimiterCombinations(c *C, key, value string) {
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
