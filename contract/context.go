package contract

import "context"

type WithContext[T any] interface {
	WithContext(ctx context.Context) T
}
