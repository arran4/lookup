package lookup

import (
	"reflect"
	"testing"
)

type S struct {
	Val int
}

func TestIntersectionRegression_Pointers(t *testing.T) {
	s1 := &S{Val: 1}
	s2 := &S{Val: 1}

	// s1 != s2 (pointers differ)
	// reflect.DeepEqual(s1, s2) is true

	left := []interface{}{s1}
	right := []interface{}{s2}

	// Helper to run Intersection
	runIntersection := func(l, r []interface{}) []interface{} {
		// Mock scope and runner
		data := struct {
			Left []interface{}
		}{
			Left: l,
		}

		res := Reflect(data).Find("Left", Intersection(NewConstantor("right", r)))
		if _, ok := res.(*Invalidor); ok {
			return nil
		}

		// Extract result
		v := res.Value() // reflect.Value
		if v.Kind() != reflect.Slice {
			return nil
		}

		out := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			out[i] = v.Index(i).Interface()
		}
		return out
	}

	res := runIntersection(left, right)
	if len(res) != 1 {
		t.Errorf("Expected 1 match (DeepEqual pointers), got %d", len(res))
	} else {
		if !reflect.DeepEqual(res[0], s1) { // DeepEqual check on result
			t.Errorf("Result mismatch")
		}
	}
}
