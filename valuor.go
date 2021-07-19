package lookup

type Valuor struct {
	Pathor
}

func ValueOf(pathor Pathor) *Valuor {
	return &Valuor{
		Pathor: pathor,
	}
}

func (v *Valuor) Run(scope *Scope, position Pathor) Pathor {
	return v.Pathor
}
