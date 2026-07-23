package properties

import "testing"

func TestNilPropertiesGet(t *testing.T) {
	var p *Properties
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Get panicked: %v", r)
		}
	}()
	v, ok := p.Get("missing")
	if ok || v != "" {
		t.Fatalf("got %q ok=%v", v, ok)
	}
}
