package properties

import "testing"

func TestLoadFileEmptyPath(t *testing.T) {
	if _, err := LoadFile("  ", UTF8); err == nil {
		t.Fatal("expected error")
	}
	if _, err := LoadFile("", UTF8); err == nil {
		t.Fatal("expected error")
	}
}
