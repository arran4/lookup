package lookup

import (
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"reflect"
	"strings"
	"testing"
)

func TestReflector_Path_StructsOnly(t *testing.T) {
	type Node1 struct {
		Name string
	}
	type Node2 struct {
		Size int
	}
	type Root struct {
		Node1 Node1
		Node2 *Node2
	}
	root := &Root{
		Node1: Node1{
			Name: "asdf",
		},
		Node2: &Node2{
			Size: 324,
		},
	}

	type Each struct {
		name             string
		Path             []string
		Expecting        interface{}
		ExpectingInvalid bool
	}
	for _, each := range []*Each{
		{name: "Empty string is self", Path: []string{""}, Expecting: root},
		{name: "Search for 'Node1' expect root.Node1", Path: []string{"Node1"}, Expecting: root.Node1},
		{name: "Search for 'Node2' expect root.Node2", Path: []string{"Node2"}, Expecting: root.Node2},
		{name: "Search for 'Node1.Name' expect root.Node1.Name", Path: []string{"Node1", "Name"}, Expecting: root.Node1.Name},
		{name: "Search for 'Node2.Size' expect root.Node2.Size", Path: []string{"Node2", "Size"}, Expecting: root.Node2.Size},
		{name: "Search for 'asdf' expect Invalidor", Path: []string{"asdf"}, ExpectingInvalid: true},
	} {
		t.Run(each.name, func(t *testing.T) {
			r := Reflect(root)
			for _, p := range each.Path {
				r = r.Find(p)
			}
			result := r.Raw()
			if each.ExpectingInvalid {
				if reflect.TypeOf(r) != reflect.TypeOf(&Invalidor{}) {
					t.Errorf("failed got %v expected invalid", result)
				}
			} else {
				if each.Expecting != result {
					t.Errorf("failed got %v expected %#v", result, each.Expecting)
				}
			}

		})
	}
}

func TestReflector_Path_StructsAndTypedSlices(t *testing.T) {
	type Node1 struct {
		Name string
	}
	type Node2 struct {
		Size int
	}
	type Root struct {
		Node1 []Node1
		Node2 []*Node2
	}
	root := &Root{
		Node1: []Node1{
			{Name: "asdf"},
			{Name: "123"},
		},
		Node2: []*Node2{
			{Size: 324},
			{Size: 213},
		},
	}

	type Each struct {
		name             string
		Path             []find
		Expecting        interface{}
		ExpectingInvalid bool
	}
	for _, each := range []*Each{
		{name: "Node 1 is a list of node 1", Path: []find{{"Node1", nil}}, Expecting: root.Node1},
		{name: "Node1[0].Name is the 1st element", Path: []find{{"Node1", nil}, {"", []Runner{Index("0")}}, {"Name", nil}}, Expecting: "asdf"},
		{name: "Node1[1].Name is the 2nd element", Path: []find{{"Node1", nil}, {"", []Runner{Index("1")}}, {"Name", nil}}, Expecting: "123"},
		{name: "Node1[-1].Name is the 2nd element", Path: []find{{"Node1", nil}, {"", []Runner{Index("-1")}}, {"Name", nil}}, Expecting: "123"},
		{name: "Node1.Name is a list of name elements in node 1", Path: []find{{"Node1", nil}, {"Name", nil}}, Expecting: []string{"asdf", "123"}},
		{name: "Node2[0].Size is the 1st element", Path: []find{{"Node2", nil}, {"", []Runner{Index("0")}}, {"Size", nil}}, Expecting: 324},
		{name: "Node2[1].Size is the 2nd element", Path: []find{{"Node2", nil}, {"", []Runner{Index("1")}}, {"Size", nil}}, Expecting: 213},
		{name: "Node2[-1].Size is the 2nd element", Path: []find{{"Node2", nil}, {"", []Runner{Index("-1")}}, {"Size", nil}}, Expecting: 213},
		{name: "Node2.Size is a list of name elements in node 2", Path: []find{{"Node2", nil}, {"Size", nil}}, Expecting: []int{324, 213}},
		{name: "Node2.Name doesn't exist in the Node2 list", Path: []find{{"Node2", nil}, {"Name", nil}}, ExpectingInvalid: true},
	} {
		t.Run(each.name, func(t *testing.T) {
			r := Reflect(root)
			for _, p := range each.Path {
				r = r.Find(p.path, p.runners...)
			}
			result := r.Raw()
			if each.ExpectingInvalid {
				if reflect.TypeOf(r) != reflect.TypeOf(&Invalidor{}) {
					t.Errorf("failed got %v expected invalid", result)
				}
			} else {
				if diff := cmp.Diff(each.Expecting, result); diff != "" {
					t.Errorf("failed got %v", diff)
				}
			}

		})
	}
}

func TestReflector_Path_StructsAndTypedSlicesWithDefaults(t *testing.T) {
	type Node1 struct {
		Name string
	}
	type Node2 struct {
		Size int
	}
	type Root struct {
		Node1 []Node1
		Node2 *Node2
	}
	root := &Root{
		Node1: []Node1{
			{Name: "asdf"},
			{Name: "123"},
		},
		Node2: &Node2{
			Size: 213,
		},
	}

	type Each struct {
		name             string
		Path             []find
		Expecting        interface{}
		ExpectingInvalid bool
	}
	for _, each := range []*Each{
		{name: "Node 1 is a list of node 1", Path: []find{{"Node1", nil}}, Expecting: root.Node1},
		{name: "Node1[0].Name is the 1st element", Path: []find{{"Node1", nil}, {"", []Runner{Index("0")}}, {"Name", nil}}, Expecting: "asdf"},
		{name: "Node1[3].Name doesn't exist return default", Path: []find{{"Node1", nil}, {"", []Runner{Index("3")}}, {"Name", nil}}, Expecting: "at:Node1"},
		{name: "Node2 Exists", Path: []find{{"Node2", nil}}, Expecting: root.Node2},
		{name: "Node2.Size Exists", Path: []find{{"Node2", nil}, {"Size", nil}}, Expecting: root.Node2.Size},
		{name: "Node2.Qty doesn't Exist", Path: []find{{"Node2", nil}, {"Qty", nil}}, Expecting: "at:Node2.Qty"},
		{name: "Node3 doesn't exist return default", Path: []find{{"Node3", nil}}, Expecting: "at:Node3"},
		{name: "Node3.LKJ doesn't exist return default", Path: []find{{"Node3", nil}, {"LKJ", nil}}, Expecting: "at:Node3"},
	} {
		t.Run(each.name, func(t *testing.T) {
			r := Reflect(root)
			paths := []string{}
			for _, p := range each.Path {
				runners := append([]Runner{}, p.runners...)
				if len(p.path) > 0 {
					paths = append(paths, p.path)
				}
				runners = append(runners, NewDefault(fmt.Sprintf("at:%s", strings.Join(paths, "."))))
				r = r.Find(p.path, runners...)
			}
			result := r.Raw()
			if each.ExpectingInvalid {
				if reflect.TypeOf(r) != reflect.TypeOf(&Invalidor{}) {
					t.Errorf("failed got %v expected invalid", result)
				}
			} else {
				if diff := cmp.Diff(each.Expecting, result); diff != "" {
					t.Errorf("failed got %v", diff)
				}
			}

		})
	}
}

func ExtractPathsFromFind(path []find) (result []string) {
	result = []string{}
	for _, p := range path {
		if len(p.path) == 0 {
			continue
		}
		result = append(result, p.path)
	}
	return result
}

func TestReflector_Path_StructsAndSlicesWithUnnamedFields(t *testing.T) {
	type UnnamedField struct {
		I float32
	}
	type Node3 struct {
		UnnamedField
	}
	type Root struct {
		Node3 *[]Node3
	}
	root := &Root{
		Node3: &[]Node3{
			{UnnamedField{I: 23}},
			{UnnamedField{I: 2}},
		},
	}

	type Each struct {
		Name             string
		Path             []string
		Expecting        interface{}
		ExpectingInvalid bool
	}
	for _, each := range []*Each{
		{Name: "Node3.I is a list of name elements in node 3 in an unnamed field struct inside node 3", Path: []string{"Node3", "", "I"}, Expecting: []float32{23, 2}},
		{Name: "Node3.UnnamedField.I is a list of name elements in node 3 in an unnamed field struct inside node 3", Path: []string{"Node3", "UnnamedField", "I"}, Expecting: []float32{23, 2}},
	} {
		t.Run(each.Name, func(t *testing.T) {
			r := Reflect(root)
			for _, p := range each.Path {
				r = r.Find(p)
			}
			result := r.Raw()
			if each.ExpectingInvalid {
				if reflect.TypeOf(r) != reflect.TypeOf(&Invalidor{}) {
					t.Errorf("failed got %v expected invalid", result)
				}
			} else {
				if diff := cmp.Diff(each.Expecting, result); diff != "" {
					t.Errorf("failed got %v", diff)
				}
			}

		})
	}
}

func TestReflector_Path_StructsAndUntypedSlices(t *testing.T) {
	type Node1 struct {
		Name string
	}
	type Node2 struct {
		Size int
	}
	type UnnamedField struct {
		I float32
	}
	type Node3 struct {
		UnnamedField
	}
	type Root struct {
		NodeN []interface{}
	}
	root := &Root{
		NodeN: []interface{}{
			&Node1{Name: "asdf"},
			&Node1{Name: "123"},
			&Node2{Size: 324},
			&Node2{Size: 213},
			&Node3{UnnamedField{I: 23}},
			&Node3{UnnamedField{I: 2}},
		},
	}

	type Each struct {
		name             string
		Path             []find
		Expecting        interface{}
		ExpectingInvalid bool
	}
	for _, each := range []*Each{
		{name: "NodeN is a list of node 1", Path: []find{{"NodeN", nil}}, Expecting: root.NodeN},
		{name: "NodeN[0].Name is the 1st element", Path: []find{{"NodeN", nil}, {"", []Runner{Index("0")}}, {"Name", nil}}, Expecting: "asdf"},
		{name: "NodeN[1].Name is the 2nd element", Path: []find{{"NodeN", nil}, {"", []Runner{Index("1")}}, {"Name", nil}}, Expecting: "123"},
		{name: "NodeN.name is a list of name elements in node 1", Path: []find{{"NodeN", nil}, {"Name", nil}}, Expecting: []string{"asdf", "123"}},
		{name: "NodeN.Size is a list of name elements in node 2", Path: []find{{"NodeN", nil}, {"Size", nil}}, Expecting: []int{324, 213}},
		{name: "NodeN.Nameeee doesn't exist in the NodeN list", Path: []find{{"NodeN", nil}, {"Nameeee", nil}}, ExpectingInvalid: true},
		{name: "NodeN.I is a list of name elements in node 3 in an unnamed field struct inside node 3", Path: []find{{"NodeN", nil}, {"", nil}, {"I", nil}}, Expecting: []float32{23, 2}},
		{name: "NodeN.UnnamedField.I is a list of name elements in node 3 in an unnamed field struct inside node 3", Path: []find{{"NodeN", nil}, {"UnnamedField", nil}, {"I", nil}}, Expecting: []float32{23, 2}},
	} {
		t.Run(each.name, func(t *testing.T) {
			r := Reflect(root)
			for _, p := range each.Path {
				r = r.Find(p.path, p.runners...)
			}
			result := r.Raw()
			if each.ExpectingInvalid {
				if reflect.TypeOf(r) != reflect.TypeOf(&Invalidor{}) {
					t.Errorf("failed got %v expected invalid", result)
				}
			} else {
				if diff := cmp.Diff(each.Expecting, result); diff != "" {
					t.Logf("Actual result %#v %v", r, r)
					t.Errorf("failed got %v", diff)
				}
			}

		})
	}
}

func TestReflector_Path_NestedMaps(t *testing.T) {
	root := map[string]interface{}{
		"test1": map[int]interface{}{
			1: 234,
			2: "two",
			3: 3.0,
		},
		"test2": map[float32]interface{}{
			1.0: 234,
			2.0: "two",
			3.0: 3.0,
		},
		"test3": map[interface{}]interface{}{
			1:   234,
			"2": "two",
			3.0: 3.0,
		},
		"test4": []map[interface{}]interface{}{
			map[interface{}]interface{}{
				1:   234,
				"2": "two",
				3.0: 3.0,
				"a": 4,
			},
			map[interface{}]interface{}{
				1:   234,
				"2": "two",
				3.0: 3.0,
				"a": "four",
			},
		},
	}

	type Each struct {
		Name             string
		Path             []string
		Expecting        interface{}
		ExpectingInvalid bool
	}
	for _, each := range []*Each{
		{Name: "test1", Path: []string{"test1"}, Expecting: map[int]interface{}{
			1: 234,
			2: "two",
			3: 3.0,
		}},
		{Name: "can use int keys", Path: []string{"test1", "3"}, Expecting: 3.0},
		{Name: "can not use float keys as int", Path: []string{"test1", "3.0"}, ExpectingInvalid: true},
		{Name: "can use float keys", Path: []string{"test2", "3.0"}, Expecting: 3.0},
		{Name: "can use int keys as float keys", Path: []string{"test2", "3"}, Expecting: 3.0},
		{Name: "can not use int keys as float key as interface{} keys", Path: []string{"test3", "3"}, ExpectingInvalid: true},
		{Name: "can not use float keys as float key as interface{} keys", Path: []string{"test3", "3.0"}, ExpectingInvalid: true},
		{Name: "array lookups over maps work single type is that type", Path: []string{"test4", "2"}, Expecting: []string{"two", "two"}},
		{Name: "array lookups over maps work multi type is interface{}", Path: []string{"test4", "a"}, Expecting: []interface{}{4, "four"}},
	} {
		t.Run(each.Name, func(t *testing.T) {
			r := Reflect(root)
			for _, p := range each.Path {
				r = r.Find(p)
			}
			result := r.Raw()
			if each.ExpectingInvalid {
				if reflect.TypeOf(r) != reflect.TypeOf(&Invalidor{}) {
					t.Errorf("failed got %v expected invalid", result)
				}
			} else {
				if diff := cmp.Diff(each.Expecting, result); diff != "" {
					t.Logf("Actual result %#v %v", r, r)
					t.Errorf("failed got %v", diff)
				}
			}

		})
	}
}

func TestReflector_Path_NestedMapsInStructsInInterfaces(t *testing.T) {
	root := map[string]interface{}{
		"test4": []struct {
			A interface{}
		}{
			{A: 4},
			{A: "four"},
		},
	}

	type Each struct {
		Name             string
		Path             []string
		Expecting        interface{}
		ExpectingInvalid bool
	}
	for _, each := range []*Each{
		{Name: "array lookups over maps work multi type is interface{}", Path: []string{"test4", "A"}, Expecting: []interface{}{4, "four"}},
	} {
		t.Run(each.Name, func(t *testing.T) {
			r := Reflect(root)
			for _, p := range each.Path {
				r = r.Find(p)
			}
			result := r.Raw()
			if each.ExpectingInvalid {
				if reflect.TypeOf(r) != reflect.TypeOf(&Invalidor{}) {
					t.Errorf("failed got %v expected invalid", result)
				}
			} else {
				if diff := cmp.Diff(each.Expecting, result); diff != "" {
					t.Logf("Actual result %#v %v", r, r)
					t.Errorf("failed got %v", diff)
				}
			}

		})
	}
}
func TestReflector_Path_MultidimensionalArray(t *testing.T) {
	root := [][]struct{ V int }{
		{
			{V: 1}, {V: 3}, {V: 5},
		}, {
			{V: 1}, {V: 3}, {V: 5},
		},
	}

	type Each struct {
		Name             string
		Path             []string
		Expecting        interface{}
		ExpectingInvalid bool
	}
	for _, each := range []*Each{
		{Name: "2d array", Path: []string{"V"}, Expecting: [][]int{{1, 3, 5}, {1, 3, 5}}},
	} {
		t.Run(each.Name, func(t *testing.T) {
			r := Reflect(root)
			for _, p := range each.Path {
				r = r.Find(p)
			}
			result := r.Raw()
			if each.ExpectingInvalid {
				if reflect.TypeOf(r) != reflect.TypeOf(&Invalidor{}) {
					t.Errorf("failed got %v expected invalid", result)
				}
			} else {
				if diff := cmp.Diff(each.Expecting, result); diff != "" {
					t.Logf("Actual result %#v %v", r, r)
					t.Errorf("failed got %v", diff)
				}
			}

		})
	}
}

type StructWithMethods struct{}

func (StructWithMethods) M1()              {}
func (StructWithMethods) M2() string       { return "hi" }
func (StructWithMethods) M3() int          { return 3 }
func (StructWithMethods) M4(i int) int     { return i }
func (StructWithMethods) M5() struct{}     { return struct{}{} }
func (StructWithMethods) M6() (int, error) { return 4, nil }
func (StructWithMethods) M7() (int, error) { return 4, errors.New("supported") }
func (*StructWithMethods) M8() string      { return "hihi" }

func TestReflector_Path_AccessingMethods(t *testing.T) {
	root := &StructWithMethods{}

	type Each struct {
		Name             string
		Path             []string
		Expecting        interface{}
		ExpectingInvalid bool
	}
	for _, each := range []*Each{
		{Name: "No return function should fail", Path: []string{"M1"}, ExpectingInvalid: true},
		{Name: "No param string return should pass", Path: []string{"M2"}, Expecting: "hi"},
		{Name: "No param int return should pass", Path: []string{"M3"}, Expecting: 3},
		{Name: "1 param int return should fail", Path: []string{"M4"}, ExpectingInvalid: true},
		{Name: "No param struct return should pass", Path: []string{"M5"}, Expecting: struct{}{}},
		{Name: "Errors are allowed as a return value", Path: []string{"M6"}, Expecting: 4},
		{Name: "Errors work as a return value", Path: []string{"M7"}, ExpectingInvalid: true},
	} {
		t.Run(each.Name, func(t *testing.T) {
			r := Reflect(root)
			for _, p := range each.Path {
				r = r.Find(p)
			}
			result := r.Raw()
			if each.ExpectingInvalid {
				if reflect.TypeOf(r) != reflect.TypeOf(&Invalidor{}) {
					t.Errorf("failed got %v expected invalid", result)
				}
			} else {
				if diff := cmp.Diff(each.Expecting, result); diff != "" {
					t.Logf("Actual result %#v %v", r, r)
					t.Errorf("failed got %v", diff)
				}
			}

		})
	}
}
func TestReflector_Path_LambdaFunc(t *testing.T) {
	root := func() interface{} {
		return struct {
			Name string
		}{
			Name: "Mr Complex",
		}
	}

	type Each struct {
		name             string
		Path             []string
		Expecting        interface{}
		ExpectingInvalid bool
	}
	for _, each := range []*Each{
		{name: "We should be able to get the result through the function", Path: []string{"Name"}, Expecting: "Mr Complex"},
	} {
		t.Run(each.name, func(t *testing.T) {
			r := Reflect(root)
			for _, p := range each.Path {
				r = r.Find(p)
			}
			result := r.Raw()
			if each.ExpectingInvalid {
				if reflect.TypeOf(r) != reflect.TypeOf(&Invalidor{}) {
					t.Errorf("failed got %v expected invalid", result)
				}
			} else {
				if diff := cmp.Diff(each.Expecting, result); diff != "" {
					t.Logf("Actual result %#v %v", r, r)
					t.Errorf("failed got %v", diff)
				}
			}

		})
	}
}
func TestReflector_Path_LambdaFuncInAnArray(t *testing.T) {
	root := []interface{}{
		func() interface{} {
			return struct {
				Name string
			}{
				Name: "Mr Complex",
			}
		}}

	type Each struct {
		Name             string
		Path             []string
		Expecting        interface{}
		ExpectingInvalid bool
	}
	for _, each := range []*Each{
		{Name: "We should be able to get the result through the function", Path: []string{"Name"}, Expecting: []string{"Mr Complex"}},
	} {
		t.Run(each.Name, func(t *testing.T) {
			r := Reflect(root)
			for _, p := range each.Path {
				r = r.Find(p)
			}
			result := r.Raw()
			if each.ExpectingInvalid {
				if reflect.TypeOf(r) != reflect.TypeOf(&Invalidor{}) {
					t.Errorf("failed got %v expected invalid", result)
				}
			} else {
				if diff := cmp.Diff(each.Expecting, result); diff != "" {
					t.Logf("Actual result %#v %v", r, r)
					t.Errorf("failed got %v", diff)
				}
			}

		})
	}
}
func TestReflector_Path_StructFuncInAnArray(t *testing.T) {
	root := []interface{}{
		StructWithMethods{},
		&StructWithMethods{},
	}

	type Each struct {
		Name             string
		Path             []string
		Expecting        interface{}
		ExpectingInvalid bool
	}
	for _, each := range []*Each{
		{Name: "We should be able to get the result through the function", Path: []string{"M2"}, Expecting: []string{"hi", "hi"}},
		{Name: "We should be able to get the result through the function even if a pointer is required", Path: []string{"M8"}, Expecting: []string{"hihi"}},
	} {
		t.Run(each.Name, func(t *testing.T) {
			r := Reflect(root)
			for _, p := range each.Path {
				r = r.Find(p)
			}
			result := r.Raw()
			if each.ExpectingInvalid {
				if reflect.TypeOf(r) != reflect.TypeOf(&Invalidor{}) {
					t.Errorf("failed got %v expected invalid", result)
				}
			} else {
				if diff := cmp.Diff(each.Expecting, result); diff != "" {
					t.Logf("Actual result %#v %v", r, r)
					t.Errorf("failed got %v", diff)
				}
			}

		})
	}
}
