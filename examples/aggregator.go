package examples

//go:generate gomocker gen --force Aggregator

type Aggregator interface {
	SumInt(vals ...int) int
}
