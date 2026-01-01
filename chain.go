package lookup

type chainFunc struct {
	first  Runner
	second Runner
}

func (c *chainFunc) Run(scope *Scope) Pathor {
	res := c.first.Run(scope)
	// If the result is invalid, usually we stop navigation, unless we want to handle errors?
	// But in JSONata, if step 1 is null, step 2 is usually not executed or returns null.
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
