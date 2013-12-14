// Copyright 2013 Frank Schroeder. All rights reserved. MIT licensed.

package properties

import (
	"fmt"
	"strings"
	"testing"

	. "github.scm.corp.ebay.com/ecg-marktplaats/cas-go/third_party/launchpad.net/gocheck"
)

func Test(t *testing.T) { TestingT(t) }

type LoadSuite struct{}

var _ = Suite(&LoadSuite{})

func (l *LoadSuite) TestKeyWithEmptyValue(c *C) {
	testAllDelimiterCombinations(c, "key", "")
}

func (l *LoadSuite) TestOneKeyValue(c *C) {
	testAllDelimiterCombinations(c, "key", "value")
}

func (l *LoadSuite) TestTwoKeysAndValues(c *C) {
	testKeyValue(c, "key1=value1\nkey2=value2", "key1", "value1", "key2", "value2")
}

func (l *LoadSuite) TestWithBlankLines(c *C) {
	testKeyValue(c, "\n\nkey=value\n\n", "key", "value")
}

func (l *LoadSuite) TestKeyWithWhitespacePrefix(c *C) {
	testKeyValue(c, " key=value", "key", "value")
	testKeyValue(c, "\fkey=value", "key", "value")
	testKeyValue(c, "\tkey=value", "key", "value")
	testKeyValue(c, " \f\tkey=value", "key", "value")
}

func (l *LoadSuite) TestWithComments(c *C) {
	input := `
# this is a comment
! and so is this
key1=value1
key#2=value#2
key!3=value!3
# and another one
! and the final one
`
	testKeyValue(c, input, "key1", "value1", "key#2", "value#2", "key!3", "value!3")
}

func (l *LoadSuite) TestValueWithTrailingSpaces(c *C) {
	testAllDelimiterCombinations(c, "key", "value   ")
}

func (l *LoadSuite) TestEscapedCharsInKey(c *C) {
	testKeyValue(c, "k\\ e\\:y\\= = value", "k e:y=", "value")
}

func (l *LoadSuite) TestUnicodeLiteralInKey(c *C) {
	testKeyValue(c, "key\\u2318 = value", "key⌘", "value")
}

func (l *LoadSuite) TestEscapedCharsInValue(c *C) {
	testKeyValue(c, "key = v\\ a\\:lu\\=e\\n\\r\\t", "key", "v a:lu=e\n\r\t")
}

func (l *LoadSuite) TestMultilineValue(c *C) {
	testKeyValue(c, "key = valueA,\\\n    valueB", "key", "valueA,valueB")
	testKeyValue(c, "key = valueA,\\\n\fvalueB", "key", "valueA,valueB")
	testKeyValue(c, "key = valueA,\\\n\tvalueB", "key", "valueA,valueB")
}

func (l *LoadSuite) TestFailWithPrematureEOF(c *C) {
	testError(c, "key", "premature EOF")
}

func (l *LoadSuite) TestFailWithNonISO8859_1Input(c *C) {
	testError(c, "key₡", "invalid ISO-8859-1 input")
}

func (l *LoadSuite) TestFailWithInvalidUnicodeLiteralInKey(c *C) {
	testError(c, "key\\ugh32 = value", "invalid unicode literal")
}

func BenchmarkNewPropertiesFromString(b *testing.B) {
	input := ""
	for i := 0; i < 1000; i++ {
		input += fmt.Sprintf("key%d=value%d\n", i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewPropertiesFromString(input)
	}
}

// tests all combinations of delimiters plus leading and/or trailing spaces.
func testAllDelimiterCombinations(c *C, key, value string) {
	delimiters := []string{"=", " =", "= ", " = ", ":", " :", ": ", " : "}
	for _, delim := range delimiters {
		testKeyValue(c, fmt.Sprintf("%s%s%s", key, delim, value), key, value)
		testKeyValue(c, fmt.Sprintf("%s%s%s\n", key, delim, value), key, value)
	}
}

// tests key/value pairs for a given input.
func testKeyValue(c *C, input string, keyvalues ...string) {
	p, err := NewPropertiesFromString(input)
	c.Assert(err, IsNil)
	c.Assert(p, NotNil)
	c.Assert(p.Len(), Equals, len(keyvalues)/2)
	for i := 0; i < len(keyvalues)/2; i += 2 {
		assertKeyValue(c, p, keyvalues[i], keyvalues[i+1])
	}
}

// tests whether a given input produces a given error message.
func testError(c *C, input, msg string) {
	_, err := NewPropertiesFromString(input)
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), msg), Equals, true)
}

func assertKeyValue(c *C, p *Properties, key, value string) {
	v, ok := p.Get(key)
	c.Assert(ok, Equals, true)
	c.Assert(v, Equals, value)
}
