package config

// interface assertions
var _ Decoder[any] = (DecoderFunc[any])(nil)

// A Decoder attempts to convert the contents
// of a file into a [T] type.
type Decoder[T any] interface {
	Decode(name string, data []byte) (*T, error)
}

// A DecoderFunc represents a function that can act as a full [Decoder].
type DecoderFunc[T any] func(string, []byte) (*T, error)

// Decode converts the contents of a file into an entity of a given [T]
// type.
func (df DecoderFunc[T]) Decode(name string, data []byte) (*T, error) {
	return df(name, data)
}
