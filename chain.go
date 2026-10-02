package lookup

type chainFunc struct {
	first  Runner
	second Runner
}

// Chain creates a runner that executes the first runner and then the second runner.
// It advances the scope's Position using `Scope.Next`, but leaves `Current` unchanged.
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

// NestChain creates a runner that executes the first runner and then the second runner.
// Unlike Chain, it advances both the scope's Current context and Position
// using `Scope.Nest`, ensuring relative lookups in the next stage search from the
// result of the previous stage.
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
