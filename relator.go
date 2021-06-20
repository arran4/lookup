package lookup

// find Stores a .Find()
type find struct {
	path     string
	pathOpts []PathOpt
}

// Relator allows you to do an Evaluate from a relative location
type Relator struct {
	finds        []*find
	positionName string
}

// PathOptSet allows Relator to be used as a PathOpt
func (r *Relator) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, &Evaluator{
		fi: r,
	})
}

type RelatorPathOpt interface {
	PathOpt
	Find(path string, opts ...PathOpt) *Relator
	Copy() *Relator
}

func FromHere() RelatorPathOpt {
	return &Relator{
		finds:        nil,
		positionName: "",
	}
}

func NewRelator() *Relator {
	return &Relator{
		finds:        nil,
		positionName: "",
	}
}

// Find stores a find request to be used in the relative location. Please note this doesn't alloc a new Relator use
// Copy for that.
func (r *Relator) Find(path string, opts ...PathOpt) *Relator {
	r.finds = append(r.finds, &find{
		path:     path,
		pathOpts: opts,
	})
	return r
}

// Copy produces a copy of the Relator
func (r *Relator) Copy() *Relator {
	fs := make([]*find, len(r.finds), len(r.finds))
	copy(fs, r.finds)
	return &Relator{
		finds:        fs,
		positionName: r.positionName,
	}
}

// Evaluate replays .Find() at the current location (executed internally) and then tried to determine if EvaluateNoArgs
// is implemented if so then return the results of that, otherwise returns p.Value() not IsZero if IsValid otherwise false
func (r *Relator) Evaluate(position Pathor) bool {
	p := position
	for _, f := range r.finds {
		p = p.Find(f.path, f.pathOpts...)
	}
	if pena, ok := p.(EvaluateNoArg); ok {
		return pena.Evaluate()
	}
	if p.Value().IsValid() {
		return !p.Value().IsZero()
	}
	return false
}
