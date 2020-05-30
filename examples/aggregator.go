package examples

//go:generate gomocker gen --force github.com/jauhararifin/gomocker/examples:Aggregator

type Aggregator interface {
	SumInt(vals ...int) int
}
