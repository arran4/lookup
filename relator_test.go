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
			got := tt.relator.Evaluate(tt.args.position)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Evaluate() = %v", diff)
			}
			copygot := copy.Evaluate(tt.args.position)
			if diff := cmp.Diff(copygot, tt.want); diff != "" {
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
		{name: "No lookup does nothing", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", FromHere()).Find("Value1") }, want: "abc"},
		{name: "Empty lookup does nothing", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", FromHere().Find("")).Find("Value1") }, want: "abc"},
		{name: "Rel lookup path matches real query", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", FromHere().Find("Value1")).Find("Value1") }, want: "abc"},
		{name: "Rel lookup path matches real query", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", FromHere().Find("Value1b")).Find("Value1") }, fail: true},
		{name: "Bad path in rel path causes failure in real query", resultFunc: func() Pathor { return Reflect(ds1).Find("Field0", FromHere().Find("Value9999")).Find("Value1") }, fail: true},
		{name: "Array look up returns only true", resultFunc: func() Pathor { return Reflect(ds1).Find("Field2", FromHere().Find("Value1")).Find("Value1") }, want: []bool{true, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resultFunc()
			if tt.fail {
				if _, ok := got.(error); !ok {
					t.Errorf("Failed expected error / failure / invalid got %v", got)
				}
				return
			}
			if diff := cmp.Diff(got.Raw(), tt.want); diff != "" {
				t.Errorf("Evaluate() = %v", diff)
			}
		})
	}
}
