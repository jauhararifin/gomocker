package examples

import "context"

type AddFunc func(ctx context.Context, a, b int) (sum int, err error)
