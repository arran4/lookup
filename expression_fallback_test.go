package lookup

import (
	"testing"
)

func TestFallbackPaths(t *testing.T) {
	data := map[string]interface{}{
		"a": 1,
		"c": 3,
	}

	r := Reflect(data)

	// Test case 1: Primary path exists
	res := r.Find("a", FallbackPaths("b", "c"))
	if res.Raw() != 1 {
		t.Errorf("Expected 1, got %v", res.Raw())
	}

	// Test case 2: Primary path missing, first fallback exists
	res = r.Find("b", FallbackPaths("c"))
	if res.Raw() != 3 {
		t.Errorf("Expected 3, got %v", res.Raw())
	}

	// Test case 3: Primary path missing, first fallback missing, second fallback exists
	res = r.Find("b", FallbackPaths("d", "c"))
	if res.Raw() != 3 {
		t.Errorf("Expected 3, got %v", res.Raw())
	}

	// Test case 4: All paths missing
	res = r.Find("b", FallbackPaths("d", "e"))
	if _, ok := res.(*Invalidor); !ok {
		t.Errorf("Expected Invalidor, got %T", res)
	}
}
