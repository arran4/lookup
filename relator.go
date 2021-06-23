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

func (r *Relator) Exists() *Evaluator {
	return NewEvaluator(Exists(r))
}

type RelatorPathOpt interface {
	Find(path string, opts ...PathOpt) *Relator
	Copy() *Relator
	Exists() *Evaluator
	Run(scope *Scope, position Pathor) Pathor
}

func Find(path string, opts ...PathOpt) RelatorPathOpt {
	return (&Relator{}).Find(path, opts...)
}

func This() RelatorPathOpt {
	return (&Relator{})
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

func (r *Relator) Run(scope *Scope, position Pathor) Pathor {
	p := position
	for _, f := range r.finds {
		p = p.Find(f.path, f.pathOpts...)
	}
	return p
}