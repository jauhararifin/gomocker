package examples

//go:generate gomocker gen Aggregator --package github.com/jauhararifin/gomocker/examples

type Aggregator interface {
	SumInt(vals ...int) int
}
