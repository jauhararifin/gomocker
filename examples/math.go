package examples

import "context"

//go:generate gomocker gen Math --package github.com/jauhararifin/gomocker/examples --target-package examples --output math_mock.go

type Math interface {
	Add(ctx context.Context, a, b int) (sum int, err error)
	Subtract(ctx context.Context, a, b int) (result int, err error)
}
