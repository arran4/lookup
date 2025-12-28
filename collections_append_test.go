package lookup

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestAppend(t *testing.T) {
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
			name:   "basic append",
			result: func() Pathor { return Reflect(data).Find("Numbers", Append(Array(3, 4))) },
			want:   []interface{}{1, 2, 3, 3, 4},
		},
		{
			name:   "append with empty left",
			result: func() Pathor { return Reflect(data).Find("Empty", Append(Array(1))) },
			want:   []interface{}{1},
		},
		{
			name:   "invalid append argument",
			result: func() Pathor { return Reflect(data).Find("Numbers", Append(Result("Bad"))) },
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
