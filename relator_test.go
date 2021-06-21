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
			got := tt.relator.Run(tt.args.position)
			if diff := cmp.Diff(got.Raw(), tt.want); diff != "" {
				t.Errorf("Evaluate() = %v", diff)
			}
			copygot := copy.Run(tt.args.position)
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
			{Value1: ""},
		},
	}
	tests := []struct {
		name       string
		want       interface{}
		resultFunc func() Pathor
		fail       bool
	}{
		{name: "No lookup does nothing", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", FromHere().Exists()).Find("Value1") }, want: "abc"},
		{name: "Empty lookup does nothing", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", FromHere().Find("").Exists()).Find("Value1") }, want: "abc"},
		{name: "Rel lookup path matches real query", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", FromHere().Find("Value1").Exists()).Find("Value1") }, want: "abc"},
		{name: "Rel lookup path matches real query", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", FromHere().Find("Value1b").Exists()).Find("Value1") }, fail: true},
		{name: "Bad path in rel path causes failure in real query", resultFunc: func() Pathor {
			return Reflect(ds1).Find("Field0", FromHere().Find("Value9999").Exists()).Find("Value1")
		}, fail: true},
		{name: "Array look up has a true so returns all", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2", FromHere().Find("Value1").Exists()).Find("Value1") }, want: []bool{true, true, false}},
		{name: "Array look up only returns true", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2").Find("Value1", FromHere().Exists()) }, want: []bool{true, true}},
		{name: "Array look up doesn't have a true so fails", resultFunc: func() Pathor {
			return Reflect(ds1).Find("Field2b", FromHere().Find("Value1").DoesContainNotZero()).Find("Value1")
		}, fail: true},
		{name: "Array look up a true so passes", resultFunc: func() Pathor {
			return Reflect(ds1).Find("Field2", FromHere().Find("Value1").DoesContainNotZero()).Find("Value1")
		}, want: []bool{true, true, false}},
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
			if diff := cmp.Diff(got.Raw(), tt.want); diff != "" {
				t.Errorf("Evaluate() = %v", diff)
			}
		})
	}
}
