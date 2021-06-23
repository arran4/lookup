package lookup

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

type Predicate interface {
	Run(scope *Scope, position Pathor) Pathor
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

func (p *pathExists) Evaluate(scope *Scope, pathor Pathor) (Pathor, error) {
	v := p.p.Run(scope, pathor).Value()
	if v.IsValid() {
		return pathor, nil
	}
	return nil, nil
}

type index struct {
	i interface{}
}

func (i *index) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, &Evaluator{fi: i})
}

func Index(i interface{}) *Evaluator {
	return &Evaluator{
		group: true,
		fi: &index{
			i: i,
		},
	}
}

func (i *index) Evaluate(scope *Scope, pathor Pathor) (Pathor, error) {
	switch pathor.Type().Kind() {
	case reflect.Array, reflect.Slice:
	default:
		return nil, ErrIndexOfNotArray
	}
	return evaluateType(scope, pathor, i.i)
}

func evaluateType(scope *Scope, pathor Pathor, i interface{}) (Pathor, error) {
	if i == nil {
		return nil, nil
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ip, err := interfaceToInt(i)
		if err != nil {
			return nil, err
		}
		return evaluateInt(pathor, ip)
	case reflect.String:
		simpleValue, err := regexp.Compile("^\\d+$")
		if err != nil {
			return nil, err
		}
		if simpleValue.MatchString(i.(string)) {
			ii, err := strconv.ParseInt(i.(string), 10, 64)
			if err != nil {
				return nil, err
			}
			return evaluateInt(pathor, int(ii))
		}
	case reflect.Struct, reflect.Ptr:
		switch ii := i.(type) {
		case Predicate:
			pathor := ii.Run(scope, pathor)
			return evaluateType(scope, pathor, pathor.Raw())
		case *Constantor:
			return evaluateType(scope, pathor, ii.Raw())
		default:
			return nil, ErrIndexValueNotValid
		}
	default:
		return nil, ErrIndexValueNotValid
	}
	return nil, ErrUnknownIndexMode
}

func evaluateInt(pathor Pathor, ip int) (Pathor, error) {
	if ip < 0 {
		ip = pathor.Value().Len() + ip
	}
	if ip < 0 || ip >= pathor.Value().Len() {
		return nil, ErrIndexOutOfRange
	}
	return pathor.Find(fmt.Sprintf("%d", ip)), nil
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

func (p *isZeroValue) Evaluate(scope *Scope, pathor Pathor) (Pathor, error) {
	v := p.p.Run(scope, pathor).Value()
	if v.IsValid() && v.IsZero() {
		return pathor, nil
	}
	return nil, nil
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

func (p *not) Evaluate(scope *Scope, pathor Pathor) (Pathor, error) {
	b, err := p.p.Evaluate(scope, pathor)
	if b != nil {
		return nil, err
	} else {
		return pathor, err
	}
}

type contains struct {
	value Predicate
}

func (p *contains) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, &Evaluator{fi: p})
}

func Contains(value Predicate) *Evaluator {
	return &Evaluator{
		group: true,
		fi: &contains{
			value: value,
		},
	}
}

func (p *contains) Evaluate(scope *Scope, pathor Pathor) (Pathor, error) {
	in := pathor.Value()
	v := p.value.Run(scope, pathor)
	if elementOf(v.Value(), in, nil) {
		return pathor, nil
	}
	return nil, nil
}

type in struct {
	values Predicate
}

func (p *in) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, &Evaluator{fi: p})
}

func In(predicate Predicate) *Evaluator {
	return &Evaluator{
		fi: &in{
			values: predicate,
		},
	}
}

func (p *in) Evaluate(scope *Scope, pathor Pathor) (Pathor, error) {
	v := pathor.Value()
	in := p.values.Run(scope, pathor)
	if elementOf(v, in.Value(), nil) {
		return pathor, nil
	}
	return nil, ErrValueNotIn
}
