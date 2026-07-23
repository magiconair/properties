package properties

import "testing"

func TestLoadUTF8BOM(t *testing.T) {
	// Editors on Windows often prepend a UTF-8 BOM.
	data := []byte("\xef\xbb\xbfa=1\nb=2\n")
	p, err := Load(data, UTF8)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := p.GetString("a", ""); got != "1" {
		t.Fatalf("a: got %q want 1 (keys=%v)", got, p.Keys())
	}
	if got := p.GetString("b", ""); got != "2" {
		t.Fatalf("b: got %q want 2", got)
	}
}

func TestLoadStringUTF8BOM(t *testing.T) {
	p, err := LoadString("\ufefffoo=bar")
	if err != nil {
		t.Fatalf("LoadString: %v", err)
	}
	if got := p.GetString("foo", ""); got != "bar" {
		t.Fatalf("foo: got %q want bar (keys=%v)", got, p.Keys())
	}
}
