package lookup

import (
	"errors"
	"testing"
)

var sentinelErr = errors.New("sentinel")

type errorStruct struct{}

func (errorStruct) Fail() (int, error) { return 0, sentinelErr }

func TestArrayNavigationEdges(t *testing.T) {
	arr := []int{10, 20, 30}
	tests := []struct {
		name  string
		idx   interface{}
		valid bool
	}{
		{"in range", 1, true},
		{"negative in range", -1, true},
		{"positive out of range", 5, false},
		{"negative out of range", -4, false},
		{"invalid index", "foo", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Reflect(arr).Find("", Index(tt.idx))
			_, invalid := r.(*Invalidor)
			if invalid == tt.valid {
				t.Fatalf("expected valid=%v got invalid=%v", tt.valid, invalid)
			}
		})
	}
}

func TestMapKeyConversion(t *testing.T) {
	intMap := map[int]string{1: "one", -2: "two"}
	if v := Reflect(intMap).Find("1").Raw(); v != "one" {
		t.Fatalf("expected one got %v", v)
	}
	if _, ok := Reflect(intMap).Find("bad").(*Invalidor); !ok {
		t.Fatalf("expected invalid for bad int")
	}

	boolMap := map[bool]int{true: 1, false: 0}
	if v := Reflect(boolMap).Find("true").Raw(); v != 1 {
		t.Fatalf("expected 1 got %v", v)
	}
	if _, ok := Reflect(boolMap).Find("notbool").(*Invalidor); !ok {
		t.Fatalf("expected invalid for notbool")
	}

	floatMap := map[float32]string{1.5: "v"}
	if v := Reflect(floatMap).Find("1.5").Raw(); v != "v" {
		t.Fatalf("expected v got %v", v)
	}
	if _, ok := Reflect(floatMap).Find("x").(*Invalidor); !ok {
		t.Fatalf("expected invalid for x")
	}
}

func TestErrorWrapping(t *testing.T) {
	r := Reflect(&errorStruct{}).Find("Fail")
	if !errors.Is(r.(error), sentinelErr) {
		t.Fatalf("expected wrapped sentinel error")
	}
}
