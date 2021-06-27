package lookup

type Evaluate interface {
	Evaluate(scope *Scope, position Pathor) (Pathor, error)
	Finder
}

// Evaluator wraps either a interface or a function. It uses reflection to match type as much as possible
type Evaluator struct {
	fi   Predicate
	path string
}

func (e *Evaluator) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, e)
}

func NewEvaluator(pathor Pathor) *Evaluator {
	return &Evaluator{
		fi:   pathor,
		path: "",
	}
}

func (e *Evaluator) Evaluate(scope *Scope, position Pathor) Pathor {
	if e.fi == nil {
		return position
	}
	return e.fi.Find(e.path, scope)
}

type Scope struct {
	Current Pathor
	Parent  *Scope
	Args    []Pathor
}

func (s *Scope) PathOptSet(settings *PathSettings) {
	settings.Scope = s
}

func (s *Scope) Copy() *Scope {
	return &Scope{
		Current: s.Current,
		Parent:  s.Parent,
	}
}

func (s *Scope) Nest(new Pathor, args ...Pathor) *Scope {
	return &Scope{
		Current: new,
		Parent:  s,
		Args:    args,
	}
}
