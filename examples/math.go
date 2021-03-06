package examples

import "context"

type Math interface {
	Add(ctx context.Context, a, b int) (sum int, err error)
	Subtract(ctx context.Context, a, b int) (result int, err error)
}
