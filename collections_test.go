package lookup

import (
	"github.com/google/go-cmp/cmp"
	"strings"
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

func TestRangeConstantor(t *testing.T) {
	data := []int{0, 1, 2, 3}
	r := Reflect(data).Find("", Range(Constant(1), Constant(3)))
	if diff := cmp.Diff([]int{1, 2}, r.Raw()); diff != "" {
		t.Errorf("unexpected result: %s", diff)
	}
}

func TestRangeRunnerConstantor(t *testing.T) {
	data := []string{"a", "b", "c", "d"}
	var start Runner = Constant(1)
	r := Reflect(data).Find("", Range(start, 3))
	if diff := cmp.Diff([]string{"b", "c"}, r.Raw()); diff != "" {
		t.Errorf("unexpected result: %s", diff)
	}
}

func TestIndexConstantorPath(t *testing.T) {
	type Root struct{ Arr []int }
	root := &Root{Arr: []int{0, 1, 2}}

	r := Reflect(root).Find("Arr").Find("", Index(NewConstantor("Const", 1)))
	t.Logf("raw=%v path=%s", r.Raw(), ExtractPath(r))

	if diff := cmp.Diff(1, r.Raw()); diff != "" {
		t.Errorf("value mismatch: %s", diff)
	}
	if p := ExtractPath(r); !strings.HasPrefix(p, "Arr[1]") {
		t.Errorf("path mismatch got %s", p)
	}
}

func TestContainsMap(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	r := Reflect(m).Find("", Contains(Constant(2)))
	if diff := cmp.Diff(true, r.Raw()); diff != "" {
		t.Errorf("value mismatch: %s", diff)
	}
}

func TestIntersectionMixedTypes(t *testing.T) {
	type S struct{ Val int }
	data := struct {
		Mixed []interface{}
	}{
		Mixed: []interface{}{1, "a", S{1}, S{2}},
	}

	tests := []struct {
		name   string
		result func() Pathor
		want   interface{}
	}{
		{
			name:   "intersect safe types",
			result: func() Pathor { return Reflect(data).Find("Mixed", Intersection(Array(1, "b"))) },
			want:   []interface{}{1},
		},
		{
			name:   "intersect unsafe types",
			result: func() Pathor { return Reflect(data).Find("Mixed", Intersection(Array(S{1}, S{3}))) },
			want:   []interface{}{S{1}},
		},
		{
			name:   "intersect mixed",
			result: func() Pathor { return Reflect(data).Find("Mixed", Intersection(Array(1, S{2}, "c"))) },
			want:   []interface{}{1, S{2}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result()
			if diff := cmp.Diff(tt.want, got.Raw()); diff != "" {
				t.Errorf("unexpected result: %s", diff)
			}
		})
	}
}

func TestInStruct(t *testing.T) {
	type S struct{ A, B string }
	s := S{A: "foo", B: "bar"}
	r := Reflect("bar").Find("", In(ValueOf(Reflect(s))))
	if diff := cmp.Diff(true, r.Raw()); diff != "" {
		t.Errorf("value mismatch: %s", diff)
	}
}
