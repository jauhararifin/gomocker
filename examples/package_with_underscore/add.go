package package_with_underscore

import "context"

//go:generate gomocker gen --force AddFunc,Aggregator

type AddFunc func(ctx context.Context, a, b int) (sum int, err error)

type Aggregator interface {
	SumInt(vals ...int) int
}
