package queue

import "encoding/json"

type Codec[T any] interface {
	Marshal(value T) ([]byte, error)
	Unmarshal(data []byte) (T, error)
	ContentType() string
}

type JSONCodec[T any] struct{}

func (c JSONCodec[T]) Marshal(value T) ([]byte, error) {
	return json.Marshal(value)
}

func (c JSONCodec[T]) Unmarshal(data []byte) (T, error) {
	var v T
	err := json.Unmarshal(data, &v)
	return v, err
}

func (c JSONCodec[T]) ContentType() string {
	return "application/json"
}

