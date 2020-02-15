package examples

import "context"

type Getter interface {
	Get() int
}

//go:generate gomocker gen Sum --package github.com/jauhararifin/gomocker/examples --target-package examples --output withvararg_mock.go

type Sum func(ctx context.Context, getters ...Getter) int
