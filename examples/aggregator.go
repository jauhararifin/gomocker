package examples

//go:generate gomocker gen Aggregator --package github.com/jauhararifin/gomocker/examples --target-package examples --output aggregator_mock.go

type Aggregator interface {
	SumInt(vals ...int) int
}
