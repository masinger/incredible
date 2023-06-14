package field

import "fmt"

var NoMappingError = fmt.Errorf("no mapping found")

type Mapping[T any] struct {
	Username Accessor[T]
	Password Accessor[T]
}

func (m *Mapping[T]) DefaultAccessor(field *Field) (Accessor[T], error) {
	if field == nil {
		return m.Accessor(Password)
	} else {
		return m.Accessor(*field)
	}
}

func (m *Mapping[T]) Accessor(field Field) (Accessor[T], error) {
	accessor := m.getAccessor(field)
	if accessor == nil {
		return nil, fmt.Errorf("entry field '%s' is not supported: %w", string(field), NoMappingError)
	}
	return accessor, nil
}

func (m *Mapping[T]) getAccessor(field Field) Accessor[T] {
	switch field {
	case Username:
		return m.Username
	case Password:
		fallthrough
	default:
		return m.Password
	}
}
