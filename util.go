package lookup

// HasPath is an interface used to determine if a Pathor has a Path() function
type HasPath interface {
	Path() string
}

// ExtractPath retrieves the path use because I didn't export it.
func ExtractPath(pather Pathor) string {
	if p, ok := pather.(HasPath); ok {
		return p.Path()
	}
	switch pather := pather.(type) {
	case Invalidor:
		return pather.path
	case *Invalidor:
		return pather.path
	case *Constantor:
		return pather.path
	case *Reflector:
		return pather.path
	}
	return "unknown"
}
