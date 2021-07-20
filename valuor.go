package lookup

type Valuor struct {
	Pathor
}

func ValueOf(pathor Pathor) *Valuor {
	return &Valuor{
		Pathor: pathor,
	}
}

func (v *Valuor) Run(scope *Scope) Pathor {
	return v.Pathor
}
