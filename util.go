package lookup

func ExtractPath(pather Pathor) string {
	switch pather := pather.(type) {
	case Invalidor:
		return pather.path
	case *Invalidor:
		return pather.path
	case *Reflector:
		return pather.path
	}
	return "unknown"
}
