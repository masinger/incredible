package field

type Accessor[T any] func(element T) (string, error)
