package examples

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAggregator_SumIntMocker_MockReturnDefaultValuesOnce(t *testing.T) {
	agg, m := MakeMockedAggregator()
	m.SumInt.MockDefaultsOnce()
	r := agg.SumInt(1, 2, 3, 4, 5, 6, 7, 8)
	assert.Equal(t, 0, r)
	invocs := m.SumInt.Invocations()
	assert.Equal(t, 1, len(invocs))
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6, 7, 8}, invocs[0].Inputs.Arg1)
	assert.Equal(t, 0, invocs[0].Outputs.Out1)
	in := m.SumInt.TakeOneInvocation()
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6, 7, 8}, in.Inputs.Arg1)
	assert.Equal(t, 0, in.Outputs.Out1)
}

func TestAggregator_SumIntMocker_MockReturnValuesOnce(t *testing.T) {
	agg, m := MakeMockedAggregator()
	m.SumInt.MockOutputsOnce(10)
	r := agg.SumInt(1, 2, 3, 4, 5, 6, 7, 8)
	assert.Equal(t, 10, r)
	invocs := m.SumInt.Invocations()
	assert.Equal(t, 1, len(invocs))
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6, 7, 8}, invocs[0].Inputs.Arg1)
	assert.Equal(t, 10, invocs[0].Outputs.Out1)
	in := m.SumInt.TakeOneInvocation()
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6, 7, 8}, in.Inputs.Arg1)
	assert.Equal(t, 10, in.Outputs.Out1)
}

func TestAggregator_SumIntMocker_MockOnce(t *testing.T) {
	agg, m := MakeMockedAggregator()
	m.SumInt.MockOnce(func(i ...int) int {
		s := 0
		for _, x := range i {
			s += x
		}
		return s
	})
	r := agg.SumInt(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	assert.Equal(t, 55, r)
	invocs := m.SumInt.Invocations()
	assert.Equal(t, 1, len(invocs))
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, invocs[0].Inputs.Arg1)
	assert.Equal(t, 55, invocs[0].Outputs.Out1)
	in := m.SumInt.TakeOneInvocation()
	assert.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, in.Inputs.Arg1)
	assert.Equal(t, 55, in.Outputs.Out1)
}
