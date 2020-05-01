package examples

import "context"

//go:generate gomocker gen --force AddFunc

type AddFunc func(ctx context.Context, a, b int) (sum int, err error)
