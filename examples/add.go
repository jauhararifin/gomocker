package examples

import "context"

//go:generate gomocker gen AddFunc --package github.com/jauhararifin/gomocker/examples

type AddFunc func(ctx context.Context, a, b int) (sum int, err error)
