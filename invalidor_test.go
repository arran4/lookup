package lookup

import (
	"errors"
	"testing"
)

func TestError(t *testing.T) {
	errExpected := errors.New("expected error")
	r := Reflect(1).Find("", Error(errExpected))

	if inv, ok := r.(*Invalidor); ok {
		if inv.Unwrap() != errExpected {
			t.Errorf("expected %v, got %v", errExpected, inv.Unwrap())
		}
	} else {
		t.Errorf("expected Invalidator, got %T", r)
	}
}
