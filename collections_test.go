package lookup

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestUnion(t *testing.T) {
	data := struct {
		Numbers []int
		Empty   []int
	}{
		Numbers: []int{1, 2, 3},
	}

	tests := []struct {
		name   string
		result func() Pathor
		want   interface{}
		fail   bool
	}{
		{
			name:   "basic union",
			result: func() Pathor { return Reflect(data).Find("Numbers", Union(Array(3, 4))) },
			want:   []interface{}{1, 2, 3, 4},
		},
		{
			name:   "union with empty left",
			result: func() Pathor { return Reflect(data).Find("Empty", Union(Array(1))) },
			want:   []interface{}{1},
		},
		{
			name:   "invalid union argument",
			result: func() Pathor { return Reflect(data).Find("Numbers", Union(Result("Bad"))) },
			fail:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result()
			if tt.fail {
				if _, ok := got.(*Invalidor); !ok {
					t.Errorf("expected failure, got %#v", got.Raw())
				}
				return
			}
			if diff := cmp.Diff(tt.want, got.Raw()); diff != "" {
				t.Errorf("unexpected result: %s", diff)
			}
		})
	}
}

func TestIntersection(t *testing.T) {
	data := struct {
		Numbers []int
		Empty   []int
	}{
		Numbers: []int{1, 2, 3, 2},
	}

	tests := []struct {
		name   string
		result func() Pathor
		want   interface{}
		fail   bool
	}{
		{
			name:   "basic intersection",
			result: func() Pathor { return Reflect(data).Find("Numbers", Intersection(Array(2, 3))) },
			want:   []interface{}{2, 3},
		},
		{
			name:   "scalar intersection",
			result: func() Pathor { return Reflect(data).Find("Numbers", Intersection(Constant(2))) },
			want:   []interface{}{2},
		},
		{
			name:   "no matches",
			result: func() Pathor { return Reflect(data).Find("Numbers", Intersection(Array(4))) },
			fail:   true,
		},
		{
			name:   "nil slice",
			result: func() Pathor { return Reflect(data).Find("Empty", Intersection(Array(1))) },
			fail:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result()
			if tt.fail {
				if _, ok := got.(*Invalidor); !ok {
					t.Errorf("expected failure, got %#v", got.Raw())
				}
				return
			}
			if diff := cmp.Diff(tt.want, got.Raw()); diff != "" {
				t.Errorf("unexpected result: %s", diff)
			}
		})
	}
}

func TestFirst(t *testing.T) {
	data := struct {
		Words    []string
		Numbers  []int
		WordsNil []string
	}{
		Words:   []string{"a", "b", "b", "c"},
		Numbers: []int{1, 2},
	}

	tests := []struct {
		name   string
		result func() Pathor
		want   interface{}
		fail   bool
	}{
		{
			name:   "match",
			result: func() Pathor { return Reflect(data).Find("Words", First(Equals(Constant("b")))) },
			want:   "b",
		},
		{
			name:   "no match",
			result: func() Pathor { return Reflect(data).Find("Words", First(Equals(Constant("x")))) },
			fail:   true,
		},
		{
			name:   "non collection",
			result: func() Pathor { return Reflect(data).Find("Numbers", Index(0), First(Equals(Constant(1)))) },
			fail:   true,
		},
		{
			name:   "nil slice",
			result: func() Pathor { return Reflect(data).Find("WordsNil", First(Equals(Constant("a")))) },
			fail:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result()
			if tt.fail {
				if _, ok := got.(*Invalidor); !ok {
					t.Errorf("expected failure, got %#v", got.Raw())
				}
				return
			}
			if diff := cmp.Diff(tt.want, got.Raw()); diff != "" {
				t.Errorf("unexpected result: %s", diff)
			}
		})
	}
}

func TestLast(t *testing.T) {
	data := struct {
		Words    []string
		Numbers  []int
		WordsNil []string
	}{
		Words:   []string{"a", "b", "b", "c"},
		Numbers: []int{1, 2},
	}

	tests := []struct {
		name   string
		result func() Pathor
		want   interface{}
		fail   bool
	}{
		{
			name:   "match",
			result: func() Pathor { return Reflect(data).Find("Words", Last(Equals(Constant("b")))) },
			want:   "b",
		},
		{
			name:   "no match",
			result: func() Pathor { return Reflect(data).Find("Words", Last(Equals(Constant("x")))) },
			fail:   true,
		},
		{
			name:   "non collection",
			result: func() Pathor { return Reflect(data).Find("Numbers", Index(0), Last(Equals(Constant(1)))) },
			fail:   true,
		},
		{
			name:   "nil slice",
			result: func() Pathor { return Reflect(data).Find("WordsNil", Last(Equals(Constant("a")))) },
			fail:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result()
			if tt.fail {
				if _, ok := got.(*Invalidor); !ok {
					t.Errorf("expected failure, got %#v", got.Raw())
				}
				return
			}
			if diff := cmp.Diff(tt.want, got.Raw()); diff != "" {
				t.Errorf("unexpected result: %s", diff)
			}
		})
	}
}

func TestRange(t *testing.T) {
	data := struct {
		Words []string
	}{
		Words: []string{"a", "b", "b", "c"},
	}

	tests := []struct {
		name   string
		result func() Pathor
		want   interface{}
		fail   bool
	}{
		{
			name:   "basic range",
			result: func() Pathor { return Reflect(data).Find("Words", Range(1, 3)) },
			want:   []string{"b", "b"},
		},
		{
			name:   "negative range",
			result: func() Pathor { return Reflect(data).Find("Words", Range(-3, -1)) },
			want:   []string{"b", "b"},
		},
		{
			name:   "start greater than end",
			result: func() Pathor { return Reflect(data).Find("Words", Range(3, 1)) },
			fail:   true,
		},
		{
			name:   "end out of bounds",
			result: func() Pathor { return Reflect(data).Find("Words", Range(0, 10)) },
			fail:   true,
		},
		{
			name:   "default start",
			result: func() Pathor { return Reflect(data).Find("Words", Range(nil, 2)) },
			want:   []string{"a", "b"},
		},
		{
			name:   "default end",
			result: func() Pathor { return Reflect(data).Find("Words", Range(2, nil)) },
			want:   []string{"b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result()
			if tt.fail {
				if _, ok := got.(*Invalidor); !ok {
					t.Errorf("expected failure, got %#v", got.Raw())
				}
				return
			}
			if diff := cmp.Diff(tt.want, got.Raw()); diff != "" {
				t.Errorf("unexpected result: %s", diff)
			}
		})
	}
}
