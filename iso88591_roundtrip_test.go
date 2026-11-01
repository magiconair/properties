package properties

import (
	"bytes"
	"testing"
)

// TestWriteIso88591Roundtrip covers the ISO-8859-1 writer: every rune in
// U+0080..U+00FF has to be written as a single byte, otherwise Load with
// ISO_8859_1 (which maps each byte straight to a rune) reads the UTF-8
// sequence back as garbage.
func TestWriteIso88591Roundtrip(t *testing.T) {
	p := NewProperties()
	p.Set("café", "naïve café")

	var buf bytes.Buffer
	if _, err := p.Write(&buf, ISO_8859_1); err != nil {
		t.Fatal(err)
	}

	if bytes.Contains(buf.Bytes(), []byte{0xC3}) {
		t.Fatalf("ISO-8859-1 output contains a UTF-8 multibyte sequence: %q", buf.String())
	}

	got, err := Load(buf.Bytes(), ISO_8859_1)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := got.Get("café"); v != "naïve café" {
		t.Fatalf("roundtrip corrupted the value: got %q", v)
	}
}
