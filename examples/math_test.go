package examples

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMathMocked_Add(t *testing.T) {
	mocker, math := NewMockedMath(t)

	mocker.Add.MockReturnValuesOnce(10, nil)
	sum, err := math.Add(context.Background(), 17, 19)
	assert.Nil(t, err, "wrong error")
	assert.Equal(t, 10, sum, "wrong sum")

	assert.Equal(t, 1, len(mocker.Add.Invocations()))
	invoc := mocker.Add.TakeOneInvocation()
	assert.Equal(t, context.Background(), invoc.Parameters.arg1, "wrong context")
	assert.Equal(t, 17, invoc.Parameters.arg2, "wrong first integer")
	assert.Equal(t, 19, invoc.Parameters.arg3, "wrong second integer")
	assert.Equal(t, 10, invoc.Returns.r1, "wrong first returns")
	assert.Equal(t, nil, invoc.Returns.r2, "wrong second returns")
}

