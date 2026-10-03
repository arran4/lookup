package lookup_test

import (
	"errors"
	"fmt"
	"github.com/arran4/lookup"
	"reflect"
	"testing"
)

type customPathor struct {
	err error
}

func (c *customPathor) Find(path string, opts ...lookup.Runner) lookup.Pathor {
	return lookup.NewInvalidor("custom", c.err)
}
func (c *customPathor) Type() reflect.Type                     { return reflect.TypeOf("") }
func (c *customPathor) Value() reflect.Value                   { return reflect.Value{} }
func (c *customPathor) AsString() (string, error)              { return "", c.err }
func (c *customPathor) AsInt() (int, error)                    { return 0, c.err }
func (c *customPathor) AsFloat() (float64, error)              { return 0, c.err }
func (c *customPathor) AsBool() (bool, error)                  { return false, c.err }
func (c *customPathor) AsSlice() ([]interface{}, error)        { return nil, c.err }
func (c *customPathor) AsMap() (map[string]interface{}, error) { return nil, c.err }
func (c *customPathor) AsPtr() (interface{}, error)            { return nil, c.err }
func (c *customPathor) Interface() (interface{}, error)        { return nil, c.err }
func (c *customPathor) String() string                         { return "" }
func (c *customPathor) Raw() interface{}                       { return nil }
func (c *customPathor) RawAsInterfaceSlice() []interface{}     { return nil }

type dummyStruct struct {
	A int
}

func TestErrorsContract(t *testing.T) {
	t.Run("Reflector missing map key satisfies errors.Is(err, ErrNoSuchPath)", func(t *testing.T) {
		r := lookup.Reflect(map[string]interface{}{"a": 1})
		res := r.Find("b")
		err := res.(*lookup.Invalidor).Unwrap()
		if !errors.Is(err, lookup.ErrNoSuchPath) {
			t.Errorf("expected ErrNoSuchPath, got %v", err)
		}
	})

	t.Run("Reflector missing struct field satisfies errors.Is(err, ErrNoSuchPath)", func(t *testing.T) {
		r := lookup.Reflect(dummyStruct{A: 1})
		res := r.Find("b")
		err := res.(*lookup.Invalidor).Unwrap()
		if !errors.Is(err, lookup.ErrNoSuchPath) {
			t.Errorf("expected ErrNoSuchPath, got %v", err)
		}
	})

	t.Run("traversal through a scalar satisfies errors.Is(err, ErrNotNavigable)", func(t *testing.T) {
		r := lookup.Reflect(42)
		res := r.Find("b")
		err := res.(*lookup.Invalidor).Unwrap()
		if !errors.Is(err, lookup.ErrNotNavigable) {
			t.Errorf("expected ErrNotNavigable, got %v", err)
		}
	})

	t.Run("Simpleor and Reflector give the same sentinel classification for equivalent missing/scalar navigation", func(t *testing.T) {
		s := lookup.Simple(map[string]interface{}{"a": 1})
		res1 := s.Find("b")
		err1 := res1.(*lookup.Invalidor).Unwrap()
		if !errors.Is(err1, lookup.ErrNoSuchPath) {
			t.Errorf("Simpleor missing map key: expected ErrNoSuchPath, got %v", err1)
		}

		s2 := lookup.Simple(42)
		res2 := s2.Find("b")
		err2 := res2.(*lookup.Invalidor).Unwrap()
		if !errors.Is(err2, lookup.ErrNotNavigable) {
			t.Errorf("Simpleor traversal through scalar: expected ErrNotNavigable, got %v", err2)
		}
	})

	t.Run("Interface/custom-backed Pathor returning a genuine error preserves that error", func(t *testing.T) {
		customErr := errors.New("my custom error")
		p := &customPathor{err: customErr}
		res := p.Find("b")
		err := res.(*lookup.Invalidor).Unwrap()
		if !errors.Is(err, customErr) {
			t.Errorf("expected %v, got %v", customErr, err)
		}
	})

	t.Run("Interface/custom-backed Pathor with deliberate ErrNoSuchPath remains classifiable as missing", func(t *testing.T) {
		customErr := fmt.Errorf("wrapped error: %w", lookup.ErrNoSuchPath)
		p := &customPathor{err: customErr}
		res := p.Find("b")
		err := res.(*lookup.Invalidor).Unwrap()
		if !errors.Is(err, lookup.ErrNoSuchPath) {
			t.Errorf("expected ErrNoSuchPath, got %v", err)
		}
	})

	t.Run("a real method error remains that error and does not become ErrNoSuchPath or ErrNotNavigable", func(t *testing.T) {
		m := lookup.Reflect(map[string]interface{}{"a": 1})

		errStr, err := m.Find("a").AsString()
		_ = errStr
		if err == nil {
			t.Fatalf("expected error")
		}
		if errors.Is(err, lookup.ErrNoSuchPath) || errors.Is(err, lookup.ErrNotNavigable) {
			t.Errorf("did not expect ErrNoSuchPath or ErrNotNavigable, got %v", err)
		}
	})
}
