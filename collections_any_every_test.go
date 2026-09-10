package lookup

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAnyAndEvery(t *testing.T) {
	tests := []struct {
		name   string
		runner Runner
		want   interface{}
		fail   bool
	}{
		// Any and Every with all true
		{
			name:   "Any all true",
			runner: Any(Constant([]bool{true, true, true})),
			want:   true,
		},
		{
			name:   "Every all true",
			runner: Every(Constant([]bool{true, true, true})),
			want:   true,
		},
		// Any and Every with mixed
		{
			name:   "Any mixed",
			runner: Any(Constant([]bool{true, false, true})),
			want:   true,
		},
		{
			name:   "Every mixed",
			runner: Every(Constant([]bool{true, false, true})),
			want:   false,
		},
		// Any and Every with all false
		{
			name:   "Any all false",
			runner: Any(Constant([]bool{false, false, false})),
			want:   false,
		},
		{
			name:   "Every all false",
			runner: Every(Constant([]bool{false, false, false})),
			want:   false,
		},
		// Any and Every with empty slice
		{
			name:   "Any empty slice",
			runner: Any(Constant([]bool{})),
			want:   false,
		},
		{
			name:   "Every empty slice",
			runner: Every(Constant([]bool{})),
			want:   true,
		},
		// Any and Every with fixed array
		{
			name:   "Any fixed array all true",
			runner: Any(Constant([3]bool{true, true, true})),
			want:   true,
		},
		{
			name:   "Every fixed array all true",
			runner: Every(Constant([3]bool{true, true, true})),
			want:   true,
		},
		{
			name:   "Any fixed array mixed",
			runner: Any(Constant([3]bool{true, false, true})),
			want:   true,
		},
		{
			name:   "Every fixed array mixed",
			runner: Every(Constant([3]bool{true, false, true})),
			want:   false,
		},
		// Any and Every with nil slice
		{
			name:   "Any nil slice",
			runner: Any(Constant([]bool(nil))),
			fail:   true,
		},
		{
			name:   "Every nil slice",
			runner: Every(Constant([]bool(nil))),
			fail:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reflect(nil).Find("", tt.runner)
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
