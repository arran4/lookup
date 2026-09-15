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

func TestInvalidor_NilErrorSafety(t *testing.T) {
	// 1. nil-error construction is safe
	inv := NewInvalidor("test.path", nil)

	// 2. nil-error Error() is safe
	if inv.Error() == "" {
		t.Errorf("Error() returned empty string for nil error")
	}

	if !errors.Is(inv, ErrEvalFail) {
		t.Errorf("Expected nil error to be normalized to ErrEvalFail")
	}

	// 3. Find() remains safe and updates/preserves the path as expected
	found := inv.Find("subpath")
	if found.(*Invalidor).Path() != "test.path.subpath" {
		t.Errorf("Expected path 'test.path.subpath', got %v", found.(*Invalidor).Path())
	}

	// 4. each relevant As* path returns a sensible error rather than panicking
	_, err := inv.AsString()
	if err == nil {
		t.Errorf("AsString() expected error, got nil")
	}
	_, err = inv.AsInt()
	if err == nil {
		t.Errorf("AsInt() expected error, got nil")
	}
	_, err = inv.AsBool()
	if err == nil {
		t.Errorf("AsBool() expected error, got nil")
	}
	_, err = inv.AsFloat()
	if err == nil {
		t.Errorf("AsFloat() expected error, got nil")
	}
	_, err = inv.AsSlice()
	if err == nil {
		t.Errorf("AsSlice() expected error, got nil")
	}
	_, err = inv.AsMap()
	if err == nil {
		t.Errorf("AsMap() expected error, got nil")
	}
	_, err = inv.AsPtr()
	if err == nil {
		t.Errorf("AsPtr() expected error, got nil")
	}

	// formatting check
	_ = inv.Error()

	// raw checks
	if inv.Raw() != nil {
		t.Errorf("Raw() should be nil")
	}
	if inv.RawAsInterfaceSlice() != nil {
		t.Errorf("RawAsInterfaceSlice() should be nil")
	}
	if inv.Value().IsValid() {
		t.Errorf("Value() should return an invalid reflect.Value")
	}

	// 5. Error(nil) runner is safe
	scope := NewScope(Simple("test"), Simple("test"))
	errRunner := Error(nil)
	res := errRunner.Run(scope)
	if res.(*Invalidor).Path() != scope.Path() {
		t.Errorf("Expected runner to preserve path, got %v", res.(*Invalidor).Path())
	}

	if err, ok := res.(error); !ok || !errors.Is(err, ErrEvalFail) {
		t.Errorf("Expected Error(nil) runner to return ErrEvalFail, got %v", err)
	}
}

func TestInvalidor_NonNilErrorIdentity(t *testing.T) {
	// 6. non-nil wrapped error identity remains inspectable
	originalErr := errors.New("my custom error")
	inv := NewInvalidor("path", originalErr)

	if !errors.Is(inv, originalErr) {
		t.Errorf("Expected error identity to be preserved")
	}

	// via Error func
	res := Error(originalErr).Run(NewScope(Simple("test"), Simple("test")))
	if !errors.Is(res.(error), originalErr) {
		t.Errorf("Expected error identity to be preserved via Error runner")
	}
}
