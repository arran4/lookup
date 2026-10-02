package lookup

type chainFunc struct {
	first  Runner
	second Runner
}

func (c *chainFunc) Run(scope *Scope) Pathor {
	res := c.first.Run(scope)
	// If the result is invalid, we stop navigation.
	// Invalidor is how we represent errors/missing.
	if _, ok := res.(*Invalidor); ok {
		return res
	}
	return c.second.Run(scope.Next(res))
}

func Chain(first, second Runner) *chainFunc {
	return &chainFunc{
		first:  first,
		second: second,
	}
}

type nestChainFunc struct {
	first  Runner
	second Runner
}

func (c *nestChainFunc) Run(scope *Scope) Pathor {
	res := c.first.Run(scope)
	if _, ok := res.(*Invalidor); ok {
		return res
	}
	return c.second.Run(scope.Nest(res))
}

func NestChain(first, second Runner) *nestChainFunc {
	return &nestChainFunc{
		first:  first,
		second: second,
	}
}
