package examples

type Aggregator interface {
	SumInt(vals ...int) int
}
