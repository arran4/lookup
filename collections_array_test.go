package lookup

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestFirstLastRangeFixedArray(t *testing.T) {
	data := [3]string{"a", "b", "c"}

	tests := []struct {
		name   string
		result func() Pathor
		want   interface{}
	}{
		{
			name:   "First with array",
			result: func() Pathor { return Reflect(data).Find("", First(Equals(Constant("b")))) },
			want:   "b",
		},
		{
			name:   "Last with array",
			result: func() Pathor { return Reflect(data).Find("", Last(Equals(Constant("b")))) },
			want:   "b",
		},
		{
			name:   "Range with array",
			result: func() Pathor { return Reflect(data).Find("", Range(1, 3)) },
			want:   []string{"b", "c"},
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
