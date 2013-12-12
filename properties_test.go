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

func (l *LoadSuite) TestValueWithTrailingSpaces(c *C) {
	testAllDelimiterCombinations(c, "key", "value   ")
}

func (l *LoadSuite) TestEscapedCharsInKey(c *C) {
	testKeyValue(c, "k\\ e\\:y\\= = value", "k e:y=", "value")
}

func (l *LoadSuite) TestUnicodeLiteralInKey(c *C) {
	testKeyValue(c, "key\\u2318 = value", "key⌘", "value")
	testKeyValue(c, "key\\U2318 = value", "key⌘", "value")
}

func (l *LoadSuite) TestEscapedCharsInValue(c *C) {
	testKeyValue(c, "key = v\\ a\\:lu\\=e\\n\\r\\t", "key", "v a:lu=e\n\r\t")
}

func (l *LoadSuite) TestMultilineValue(c *C) {
	input := "key = valueA,\\\n    valueB"
	testKeyValue(c, input, "key", "valueA,valueB")
}

func (l *LoadSuite) TestFailWithPrematureEOF(c *C) {
	_, err := NewPropertiesFromString("key")
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "premature EOF"), Equals, true)
}

func (l *LoadSuite) TestFailWithNonISO8859_1Input(c *C) {
	_, err := NewPropertiesFromString("key₡")
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "invalid ISO-8859-1 input"), Equals, true)
}

func (l *LoadSuite) TestFailWithInvalidUnicodeLiteralInKey(c *C) {
	_, err := NewPropertiesFromString("key\\ugh32 = value")
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "invalid unicode literal"), Equals, true)
}

// tests all combinations of delimiters plus leading and/or trailing spaces.
func testAllDelimiterCombinations(c *C, key, value string) {
	delimiters := []string{"=", " =", "= ", " = ", ":", " :", ": ", " : "}
	for _, delim := range delimiters {
		testKeyValue(c, fmt.Sprintf("%s%s%s", key, delim, value), key, value)
		testKeyValue(c, fmt.Sprintf("%s%s%s\n", key, delim, value), key, value)
	}
}

// tests a single key/value combination for a given input.
func testKeyValue(c *C, input, key, value string) {
	// fmt.Printf("Testing '%s'\n", input)
	p, err := NewPropertiesFromString(input)
	c.Assert(err, IsNil)
	c.Assert(p, NotNil)
	c.Assert(p.Len(), Equals, 1)
	assertKeyValue(c, p, key, value)
}

func assertKeyValue(c *C, p *Properties, key, value string) {
	v, ok := p.Get(key)
	c.Assert(ok, Equals, true)
	c.Assert(v, Equals, value)
}
