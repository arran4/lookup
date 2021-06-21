package lookup

import "reflect"

type Predicate interface {
	Run(position Pathor) Pathor
}

type pathExists struct {
	p Predicate
}

func (p *pathExists) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, &Evaluator{fi: p})
}

func Exists(predicate Predicate) PathOpt {
	return &pathExists{
		p: predicate,
	}
}

func (p *pathExists) Evaluate(pathor Pathor) (bool, error) {
	v := p.p.Run(pathor).Value()
	return v.IsValid(), nil
}

type isZeroValue struct {
	p Predicate
}

func (p *isZeroValue) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, &Evaluator{fi: p})
}

func IsZero(predicate Predicate) *Evaluator {
	return NewEvaluator(&isZeroValue{
		p: predicate,
	})
}

func (p *isZeroValue) Evaluate(pathor Pathor) (bool, error) {
	v := p.p.Run(pathor).Value()
	if v.IsValid() {
		return !v.IsZero(), nil
	}
	return false, nil
}

type not struct {
	p *Evaluator
}

func (p *not) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, &Evaluator{fi: p})
}

func Not(evaluator *Evaluator) *Evaluator {
	return NewEvaluator(&not{
		p: evaluator,
	})
}

func (p *not) Evaluate(pathor Pathor) (bool, error) {
	b, err := p.p.Evaluate(pathor)
	return !b, err
}

type contains struct {
	value Pathor
}

func (p *contains) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, &Evaluator{fi: p})
}

func Contains(value Pathor) *Evaluator {
	return &Evaluator{
		fi: &contains{
			value: value,
		},
	}
}

func (p *contains) Evaluate(pathor Pathor) (bool, error) {
	in := pathor.Value()
	v := p.value.Value()
	return elementOf(v, in, nil), nil
}

type in struct {
	values Pathor
}

func (p *in) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, &Evaluator{fi: p})
}

func In(inValues Pathor) *Evaluator {
	return &Evaluator{
		fi: &in{
			values: inValues,
		},
	}
}

func (p *in) Evaluate(pathor Pathor) (bool, error) {
	v := pathor.Value()
	in := p.values.Value()
	return elementOf(v, in, nil), nil
}

func elementOf(v reflect.Value, in reflect.Value, pv *reflect.Value) bool {
	if !in.IsValid() {
		return false
	}
	if !in.IsValid() {
		return false
	}
	switch in.Kind() {
	//case reflect.Bool:
	//case reflect.Int:
	//case reflect.Int8:
	//case reflect.Int16:
	//case reflect.Int32:
	//case reflect.Int64:
	//case reflect.Uint:
	//case reflect.Uint8:
	//case reflect.Uint16:
	//case reflect.Uint32:
	//case reflect.Uint64:
	//case reflect.Uintptr:
	//case reflect.Float32:
	//case reflect.Float64:
	//case reflect.Complex64:
	//case reflect.Complex128:
	case reflect.Array:
		for i := 0; i < in.Len(); i++ {
			f := in.Index(i)
			if reflect.DeepEqual(v.Interface(), f.Interface()) {
				return true
			}
		}
	//case reflect.Chan:
	case reflect.Func:
		var r Pathor
		r = runMethod(in, "")
		return elementOf(r.Value(), in, nil)
	case reflect.Map:
		for _, k := range in.MapKeys() {
			f := in.MapIndex(k)
			if reflect.DeepEqual(v.Interface(), f.Interface()) {
				return true
			}
		}
	case reflect.Ptr:
		return elementOf(v.Elem(), in.Elem(), &v)
	case reflect.Slice:
		for i := 0; i < in.Len(); i++ {
			f := in.Index(i)
			if reflect.DeepEqual(v.Interface(), f.Interface()) {
				return true
			}
		}
	//case reflect.String:
	case reflect.Struct:
		for i := 0; i < in.NumField(); i++ {
			f := in.Field(i)
			if reflect.DeepEqual(v.Interface(), f.Interface()) {
				return true
			}
		}
		for i := 0; i < in.NumMethod(); i++ {
			var f reflect.Value
			if pv == nil {
				f = v.Method(i)
			} else {
				f = pv.Method(i)
			}
			fr := runMethod(f, "")
			if elementOf(fr.Value(), in, nil) {
				return true
			}
		}

	//case reflect.UnsafePointer:
	default:
		return reflect.DeepEqual(in.Interface(), in.Interface())
	}
	return false
}
