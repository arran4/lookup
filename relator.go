package lookup

// find Stores a .Find()
type find struct {
	path    string
	runners []Runner
}

type relationType int

const (
	relationTypeCurrent relationType = iota
	relationTypeParent
	relationTypePosition
)

// Relator allows you to do an Evaluate from a relative location
type Relator struct {
	finds        []*find
	positionName string
	relationType relationType
}

func Find(path string, opts ...Runner) *Relator {
	return (&Relator{}).Find(path, opts...)
}

func This() *Relator {
	return &Relator{}
}

func Result(path string, opts ...Runner) *Relator {
	return (&Relator{
		relationType: relationTypePosition,
	}).Find(path, opts...)
}

func NewRelator() *Relator {
	return &Relator{
		finds:        nil,
		positionName: "",
	}
}

// Find stores a find request to be used in the relative location. Please note this doesn't alloc a new Relator use
// Copy for that.
func (r *Relator) Find(path string, opts ...Runner) *Relator {
	r.finds = append(r.finds, &find{
		path:    path,
		runners: opts,
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
		relationType: r.relationType,
	}
}

func (r *Relator) Run(scope *Scope) Pathor {
	var p Pathor
	switch r.relationType {
	case relationTypeParent:
		if scope.Parent != nil {
			p = scope.Parent.Current
			break
		}
		fallthrough
	case relationTypeCurrent:
		p = scope.Current
	case relationTypePosition:
		p = scope.Position
	}
	for _, f := range r.finds {
		p = p.Find(f.path, f.runners...)
	}
	return p
}
