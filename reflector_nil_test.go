package lookup

import (
	"errors"
	"testing"
)

func TestReflectNilTypeMethodsDoNotPanic(t *testing.T) {
	p := Reflect(nil)

	assertNoPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("%s panicked for nil reflector: %v", name, r)
			}
		}()
		fn()
	}

	assertNoPanic("IsNil", func() {
		if !p.IsNil() {
			t.Fatalf("expected IsNil() to be true")
		}
	})

	assertNoPanic("IsString", func() {
		if p.IsString() {
			t.Fatalf("expected IsString() to be false")
		}
	})

	assertNoPanic("AsString", func() {
		if _, err := p.AsString(); err == nil || !errors.Is(err, ErrNotString) {
			t.Fatalf("expected ErrNotString, got %v", err)
		}
	})

	assertNoPanic("Type", func() {
		if got := p.Type(); got != nil {
			t.Fatalf("expected nil type for nil reflector, got %v", got)
		}
	})
}
