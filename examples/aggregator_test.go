package examples
//
//import (
//	"github.com/stretchr/testify/assert"
//	"testing"
//)
//
//func TestAggregator_SumIntMocker_MockReturnDefaultValuesOnce(t *testing.T) {
//	m, agg := NewMockedAggregator(t)
//	m.SumInt.MockReturnDefaultValuesOnce()
//	r := agg.SumInt(1, 2, 3, 4, 5, 6, 7, 8)
//	assert.Equal(t, 0, r)
//	invocs := m.SumInt.Invocations()
//	assert.Equal(t, 1, len(invocs))
//	assert.ElementsMatch(t, []int{1,2,3,4,5,6,7,8}, invocs[0].Parameters.Arg1)
//	assert.Equal(t, 0, invocs[0].Returns.R1)
//	in := m.SumInt.TakeOneInvocation()
//	assert.ElementsMatch(t, []int{1,2,3,4,5,6,7,8}, in.Parameters.Arg1)
//	assert.Equal(t, 0, in.Returns.R1)
//}
//
//func TestAggregator_SumIntMocker_MockReturnValuesOnce(t *testing.T) {
//	m, agg := NewMockedAggregator(t)
//	m.SumInt.MockReturnValuesOnce(10)
//	r := agg.SumInt(1, 2, 3, 4, 5, 6, 7, 8)
//	assert.Equal(t, 10, r)
//	invocs := m.SumInt.Invocations()
//	assert.Equal(t, 1, len(invocs))
//	assert.ElementsMatch(t, []int{1,2,3,4,5,6,7,8}, invocs[0].Parameters.Arg1)
//	assert.Equal(t, 10, invocs[0].Returns.R1)
//	in := m.SumInt.TakeOneInvocation()
//	assert.ElementsMatch(t, []int{1,2,3,4,5,6,7,8}, in.Parameters.Arg1)
//	assert.Equal(t, 10, in.Returns.R1)
//}
//
//func TestAggregator_SumIntMocker_MockOnce(t *testing.T) {
//	m, agg := NewMockedAggregator(t)
//	m.SumInt.MockOnce(func(i ...int) int {
//		s := 0
//		for _, x := range i {
//			s += x
//		}
//		return s
//	})
//	r := agg.SumInt(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
//	assert.Equal(t, 55, r)
//	invocs := m.SumInt.Invocations()
//	assert.Equal(t, 1, len(invocs))
//	assert.ElementsMatch(t, []int{1,2,3,4,5,6,7,8,9,10}, invocs[0].Parameters.Arg1)
//	assert.Equal(t, 55, invocs[0].Returns.R1)
//	in := m.SumInt.TakeOneInvocation()
//	assert.ElementsMatch(t, []int{1,2,3,4,5,6,7,8,9,10}, in.Parameters.Arg1)
//	assert.Equal(t, 55, in.Returns.R1)
//}
