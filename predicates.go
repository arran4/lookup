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

func Exists(predicate Predicate) *Evaluator {
	return NewEvaluator(&pathExists{
		p: predicate,
	})
}

func (p *pathExists) Evaluate(pathor Pathor) bool {
	v := p.p.Run(pathor).Value()
	if v.IsValid() {
		return !v.IsZero()
	}
	return false
}

type notZeroValue struct {
	p Predicate
}

func (p *notZeroValue) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, &Evaluator{fi: p})
}

// NotZero does a simple reflect.Value().IsZero check to see if the object is zero if it is then returns false otherwise true
func NotZero(predicate Predicate) *Evaluator {
	return NewEvaluator(&notZeroValue{
		p: predicate,
	})
}

func (p *notZeroValue) Evaluate(pathor Pathor) bool {
	v := p.p.Run(pathor).Value()
	if v.IsValid() {
		return !v.IsZero()
	}
	return false
}

type containsNotZeroValue struct {
	p Predicate
}

func (p *containsNotZeroValue) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, &Evaluator{fi: p})
}

// ContainsNotZero recursively explores the object for a not zero reflect.Value -- warning no constraints
func ContainsNotZero(predicate Predicate) *Evaluator {
	return &Evaluator{
		fi: &containsNotZeroValue{
			p: predicate,
		},
	}
}

func (p *containsNotZeroValue) Evaluate(pathor Pathor) bool {
	v := p.p.Run(pathor).Value()
	return getContainsANotZero(v)
}

func getContainsANotZero(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}
	switch v.Kind() {
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
		for i := 0; i < v.Len(); i++ {
			f := v.Index(i)
			if getContainsANotZero(f) {
				return true
			}
		}
	//case reflect.Chan:
	case reflect.Func:
		r := runMethod(v, "")
		return getContainsANotZero(r.Value())
	case reflect.Map:
		for _, k := range v.MapKeys() {
			f := v.MapIndex(k)
			if getContainsANotZero(f) {
				return true
			}
		}
	case reflect.Ptr:
		return getContainsANotZero(v.Elem())
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			f := v.Index(i)
			if getContainsANotZero(f) {
				return true
			}
		}
	//case reflect.String:
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if getContainsANotZero(f) {
				return true
			}
		}
		for i := 0; i < v.NumMethod(); i++ {
			f := v.Method(i)
			if getContainsANotZero(f) {
				return true
			}
		}

	//case reflect.UnsafePointer:
	default:
		return !v.IsZero()
	}
	return false
}
