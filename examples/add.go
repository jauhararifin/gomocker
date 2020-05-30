package examples

import "context"

//go:generate gomocker gen --force github.com/jauhararifin/gomocker/examples:AddFunc

type AddFunc func(ctx context.Context, a, b int) (sum int, err error)
