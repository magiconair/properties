// Copyright 2013 Frank Schroeder. All rights reserved. MIT licensed.

package properties

import (
	"fmt"
)

func ExampleDecode() {
	buf := []byte("key = ISO-8859-1 value with unicode literal \\u2318 and umlaut ")
	buf = append(buf, 0xE4) // 0xE4 == ä
	p, _ := Decode(buf)
	v, ok := p.Get("key")
	fmt.Println(ok)
	fmt.Println(v)
	// Output:
	// true
	// ISO-8859-1 value with unicode literal ⌘ and umlaut ä
}

func ExampleDecodeFromString() {
	p, _ := DecodeFromString("key = UTF-8 value with unicode character ⌘ and umlaut ä")
	v, ok := p.Get("key")
	fmt.Println(ok)
	fmt.Println(v)
	// Output:
	// true
	// UTF-8 value with unicode character ⌘ and umlaut ä
}
