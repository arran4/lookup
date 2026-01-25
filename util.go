package lookup

import (
	"fmt"
	"reflect"
)

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
	case *Invalidor:
		return pather.path
	case *Constantor:
		return pather.path
	case *Reflector:
		return pather.path
	case *Interfaceor:
		return pather.path
	}
	return "unknown"
}

type CustomPath interface {
	Path(previousPath string, findPath string) string
}

func PathBuilder(path string, r Pathor, cp CustomPath) string {
	p := ExtractPath(r)
	if cp != nil {
		return cp.Path(p, path)
	}
	if len(p) > 0 {
		p = p + "." + path
	} else {
		p = path
	}
	return p
}

func interfaceToFloat(i interface{}) (float64, error) {
	if i == nil {
		return 0, fmt.Errorf("nil")
	}
	switch v := i.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case reflect.Value:
		if v.CanInterface() {
			return interfaceToFloat(v.Interface())
		}
	}
	return 0, fmt.Errorf("not a number")
}

func ToInt(i interface{}) (int64, bool) {
	if i == nil {
		return 0, false
	}
	switch v := i.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return int64(v), true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case reflect.Value:
		if v.CanInterface() {
			return ToInt(v.Interface())
		}
	}
	return 0, false
}
