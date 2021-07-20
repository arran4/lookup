package lookup

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestRelator_Evaluate(t *testing.T) {
	var obj1 = struct {
		F2 bool
		F1 bool
	}{
		F1: true,
		F2: false,
	}
	var c1 = NewConstantor("", true)
	var c2 = NewConstantor("", false)
	var c3 = NewConstantor("Const", true)
	var c4 = NewConstantor("Const", false)
	type args struct {
		position Pathor
	}
	tests := []struct {
		name    string
		relator *Relator
		args    args
		want    bool
	}{
		{name: "Field true", relator: NewRelator().Find("F1"), args: args{position: Reflect(obj1)}, want: true},
		{name: "Field false", relator: NewRelator().Find("F2"), args: args{position: Reflect(obj1)}, want: false},
		{name: "Constant true", relator: NewRelator(), args: args{position: c1}, want: true},
		{name: "Constant false", relator: NewRelator(), args: args{position: c2}, want: false},
		{name: "Constant with path true", relator: NewRelator(), args: args{position: c3}, want: true},
		{name: "Constant with path false", relator: NewRelator(), args: args{position: c4}, want: false},
		{name: "Constant Fields right with path true", relator: NewRelator().Find("Const"), args: args{position: c3}, want: true},
		{name: "Constant Fields right with path false", relator: NewRelator().Find("Const"), args: args{position: c4}, want: false},
		{name: "Constant Fields wrong with path true", relator: NewRelator().Find("Inconsistent"), args: args{position: c3}, want: true},
		{name: "Constant Fields wrong with path false", relator: NewRelator().Find("Inconsistent"), args: args{position: c4}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copy := tt.relator.Copy()
			got := tt.relator.Run(NewScope(tt.args.position, tt.args.position))
			if diff := cmp.Diff(got.Raw(), tt.want); diff != "" {
				t.Errorf("Evaluate() = %v", diff)
			}
			copygot := copy.Run(NewScope(tt.args.position, tt.args.position))
			if diff := cmp.Diff(copygot.Raw(), tt.want); diff != "" {
				t.Errorf("Evaluate() = %v", diff)
			}
		})
	}
}

func TestRelator_FromHere(t *testing.T) {
	type BoolField struct {
		Value1 bool
	}
	type StringField struct {
		Value1 string
	}
	ds1 := struct {
		Field0  StringField
		Field1  BoolField
		Field1b BoolField
		Field2  []BoolField
		Field2b []BoolField
		Field3  []StringField
		Field4  interface{}
		Field5  []string
		Field5a []string
	}{
		Field0:  StringField{Value1: "abc"},
		Field1:  BoolField{Value1: true},
		Field1b: BoolField{Value1: false},
		Field2: []BoolField{
			{Value1: true},
			{Value1: true},
			{Value1: false},
		},
		Field2b: []BoolField{
			{Value1: false},
			{Value1: false},
			{Value1: false},
		},
		Field3: []StringField{
			{Value1: "asdf"},
			{Value1: "This"},
			{Value1: ""},
		},
		Field5: []string{
			"Once",
			"Again",
			"This",
		},
		Field5a: []string{
			"Once",
			"Again",
		},
	}
	tests := []struct {
		name       string
		want       interface{}
		resultFunc func() Pathor
		fail       bool
	}{
		{name: "No lookup does nothing", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0").Find("Value1") }, want: "abc"},
		{name: "Empty lookup does nothing", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", Match(Find(""))).Find("Value1") }, want: "abc"},
		{name: "Rel lookup path matches real query", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", Match(Find("Field0").Find("Value1"))).Find("Value1") }, want: "abc"},
		{name: "Rel lookup path fails if it fails", resultFunc: func() Pathor {
			return Reflect(ds1).Find("Field0", Match(Find("Field0").Find("Value1b"))).Find("Value1")
		}, fail: true},
		{name: "Result lookup path matches real query", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", Match(Result("Value1"))).Find("Value1") }, want: "abc"},
		{name: "Result lookup path fails if it fails", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", Match(Result("Value1b"))).Find("Value1") }, fail: true},
		{name: "Bad path in rel path causes failure in real query", resultFunc: func() Pathor {
			return Reflect(ds1).Find("Field0", Match(Find("Value9999"))).Find("Value1")
		}, fail: true},
		{name: "Array look up has a true so returns all", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2", Match(Find("Value1"))).Find("Value1") }, want: []bool{true, true, false}},
		{name: "Array filter look up only returns [true, true]", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2").Find("Value1", Filter(This())) }, want: []bool{true, true}},
		{name: "Array iszero filter look up only returns [true, true]", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2").Find("Value1", Filter(Not(IsZero(This())))) }, want: []bool{true, true}},
		{name: "Array Field2 match look up only returns [true,true,false]", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2").Find("Value1", Match(IsZero(This()))) }, want: []bool{true, true, false}},
		{name: "Array Field2 match look up fails", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2").Find("Value1", Match(IsZero(This()))) }, want: []bool{true, true, false}},
		{name: "Array Field2b match look up fails", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2b").Find("Value1", Match(IsZero(This()))) }, fail: true},
		{name: "Array index 0 look up only returns true", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2").Find("Value1", Index(0)) }, want: true},
		{name: "Array index 1 look up only returns true", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2").Find("Value1", Index(1)) }, want: true},
		{name: "Array index 2 look up only returns false", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2").Find("Value1", Index(2)) }, want: false},
		{name: "Array index -1 look up only returns false", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2").Find("Value1", Index(-1)) }, want: false},
		{name: "Array index -2 look up only returns true", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2").Find("Value1", Index(-2)) }, want: true},
		{name: "Array index -3 look up only returns true", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2").Find("Value1", Index(-3)) }, want: true},
		{name: "In array succeeds", resultFunc: func() Pathor { return Reflect(ds1).Find("Field3").Find("Value1", Match(In(Array("This")))) }, want: []string{"This"}},
		{name: "In array fails", resultFunc: func() Pathor { return Reflect(ds1).Find("Field3").Find("Value1", Match(In(Array("NotThis")))) }, want: false},
		{name: "In pathor succeeds", resultFunc: func() Pathor {
			return Reflect(ds1).Find("Field3").Find("Value1", Match(In(ValueOf(Reflect(ds1).Find("Field5")))))
		}, want: []string{"This"}},
		{name: "In pathor fails", resultFunc: func() Pathor {
			return Reflect(ds1).Find("Field3").Find("Value1", Match(In(ValueOf(Reflect(ds1).Find("Field5a")))))
		}, want: false},
		{name: "Filter array succeeds", resultFunc: func() Pathor { return Reflect(ds1).Find("Field3").Find("Value1", Filter(Equals(Constant("This")))) }, want: []string{"This"}},
		{name: "Contains array succeeds", resultFunc: func() Pathor { return Reflect(ds1).Find("Field3").Find("Value1", Match(Contains(Constant("This")))) }, want: []string{"asdf", "This", ""}},
		{name: "Contains array fails", resultFunc: func() Pathor { return Reflect(ds1).Find("Field3").Find("Value1", Match(Contains(Constant("NotThis")))) }, want: false},
		{name: "Contains pathor succeeds", resultFunc: func() Pathor {
			return Reflect(ds1).Find("Field3").Find("Value1", Match(Contains(ValueOf(Reflect(ds1).Find("Field5")))))
		}, want: []string{"asdf", "This", ""}},
		{name: "Contains pathor fails", resultFunc: func() Pathor {
			return Reflect(ds1).Find("Field3").Find("Value1", Match(Contains(ValueOf(Reflect(ds1).Find("Field5b")))))
		}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resultFunc()
			if tt.fail {
				if _, ok := got.(error); !ok {
					t.Errorf("Failed expected error / failure / invalid got %v %#v", got, got.Raw())
				}
				return
			}
			if _, ok := got.(error); ok {
				t.Errorf("Failed unexpected error / failure / invalid got %v %#v", got, got.Raw())
			}
			if diff := cmp.Diff(got.Raw(), tt.want); diff != "" {
				t.Errorf("Evaluate() = %v", diff)
			}
		})
	}
}
