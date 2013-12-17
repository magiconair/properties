// Copyright 2013 Frank Schroeder. All rights reserved. MIT licensed.

package properties

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.scm.corp.ebay.com/ecg-marktplaats/cas-go/third_party/launchpad.net/gocheck"
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
	{"key1=value1\nkey2=value2", "key1", "value1", "key2", "value2"},

	// blank lines
	{"\n\nkey=value\n\n", "key", "value"}, // leading and trailing new lines

	// escaped chars
	{"k\\ e\\:y\\= = value", "k e:y=", "value"},                // escaped chars in key
	{"key = v\\ a\\:lu\\=e\\n\\r\\t", "key", "v a:lu=e\n\r\t"}, // escaped chars in value

	// unicode literals
	{"key\\u2318 = value", "key⌘", "value"}, // unicode literal in key

	// multiline values
	{"key = valueA,\\\n    valueB", "key", "valueA,valueB"},   // SPACE indent
	{"key = valueA,\\\n\f\f\fvalueB", "key", "valueA,valueB"}, // FF indent
	{"key = valueA,\\\n\t\t\tvalueB", "key", "valueA,valueB"}, // TAB indent
	{"key = valueA,\\\n \f\tvalueB", "key", "valueA,valueB"},  // mix indent

	// comments
	{"# this is a comment\n! and so is this\nkey1=value1\nkey#2=value#2\n\nkey!3=value!3\n# and another one\n! and the final one", "key1", "value1", "key#2", "value#2", "key!3", "value!3"},
}

// define error test cases in the form of
// {"input", "expected error message"}
var errorTests = [][]string{
	{"key", "premature EOF"},
	{"key\\ugh32 = value", "invalid unicode literal"},
}

// tests basic single key/value combinations with all possible whitespace, delimiter and newline permutations.
func (l *TestSuite) TestBasic(c *C) {
	testAllCombinations(c, "key", "")
	testAllCombinations(c, "key", "value")
	testAllCombinations(c, "key", "value   ")
}

func (l *TestSuite) TestComplex(c *C) {
	for i, test := range complexTests {
		printf("[C%02d] %q %q\n", i, test[0], test[1:])
		testKeyValue(c, test[0], test[1:]...)
	}
}

func (l *TestSuite) TestErrors(c *C) {
	for i, test := range errorTests {
		input, msg := test[0], test[1]
		printf("[E%02d] %q %q\n", i, input, msg)
		testError(c, input, msg)
	}
}

func BenchmarkDecoder(b *testing.B) {
	input := ""
	for i := 0; i < 1000; i++ {
		input += fmt.Sprintf("key%d=value%d\n", i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := NewDecoder(strings.NewReader(input))
		d.Decode()
	}
}

// tests all combinations of delimiters plus leading and/or trailing spaces.
func testAllCombinations(c *C, key, value string) {
	whitespace := []string{" ", "\f", "\t"}
	delimiters := []string{"", "=", ":"}
	// newlines := []string{"", "\r", "\n", "\r\n"}
	newlines := []string{"", "\n"}
	for _, dl := range delimiters {
		for _, ws1 := range whitespace {
			for _, ws2 := range whitespace {
				for _, nl := range newlines {
					input := fmt.Sprintf("%s%s%s%s%s%s", key, ws1, dl, ws2, value, nl)
					printf("%q\n", input)
					testKeyValue(c, input, key, value)
				}
			}
		}
	}
}

// tests key/value pairs for a given input.
func testKeyValue(c *C, input string, keyvalues ...string) {
	d := NewDecoder(strings.NewReader(input))
	p, err := d.Decode()
	c.Assert(err, IsNil)
	c.Assert(p, NotNil)
	c.Assert(p.Len(), Equals, len(keyvalues)/2, Commentf("Odd number of key/value pairs."))
	for i := 0; i < len(keyvalues)/2; i += 2 {
		key, value := keyvalues[i], keyvalues[i+1]
		v, ok := p.Get(key)
		c.Assert(ok, Equals, true, Commentf("No key %q for input %q", key, input))
		c.Assert(v, Equals, value, Commentf("Value %q does not match input %q", value, input))
	}
}

// tests whether a given input produces a given error message.
func testError(c *C, input, msg string) {
	d := NewDecoder(strings.NewReader(input))
	_, err := d.Decode()
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), msg), Equals, true)
}

func printf(format string, args ...interface{}) {
	if *verbose {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}
