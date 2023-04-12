package go_cache

import "context"

type Value[T any] interface {
	Value() T
	TTL() uint32
}
type Cache[S, T any] interface {
	Get(ctx context.Context, key S) (Value[T], error)
	Set(ctx context.Context, key S, value Value[T]) error
	Delete(ctx context.Context, key S) error
}
