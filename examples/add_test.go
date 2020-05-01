package examples

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddFuncMocker_MockReturnDefaultValueOnce(t *testing.T) {
	f, mocker := NewMockedAddFunc()
	mocker.MockDefaultsOnce()
	sum, err := f(context.Background(), 10, 20)
	assert.Nil(t, err, "unexpected error when calling AddFunc")
	assert.Equal(t, 0, sum, "sum should contain default value, which is zero")
}

func TestAddFuncMocker_MockReturnDefaultValueForever(t *testing.T) {
	f, mocker := NewMockedAddFunc()
	mocker.MockDefaultsForever()
	for i := 0; i < 100; i++ {
		sum, err := f(context.Background(), 10, 20)
		assert.Nil(t, err, "unexpected error when calling AddFunc")
		assert.Equal(t, 0, sum, "sum should contain default value, which is zero")
	}
}

func TestAddFuncMocker_MockReturnDefaultValue(t *testing.T) {
	f, mocker := NewMockedAddFunc()
	mocker.MockDefaults(10)
	for i := 0; i < 10; i++ {
		sum, err := f(context.Background(), 10, 20)
		assert.Nil(t, err, "unexpected error when calling AddFunc")
		assert.Equal(t, 0, sum, "sum should contain default value, which is zero")
	}
}

func TestAddFuncMocker_MockReturnValueOnce(t *testing.T) {
	f, mocker := NewMockedAddFunc()
	mocker.MockOutputsOnce(20, errors.New("error_test"))
	sum, err := f(context.Background(), 10, 20)
	assert.NotNil(t, err)
	assert.Equal(t, "error_test", err.Error())
	assert.Equal(t, 20, sum)
}

func TestAddFuncMocker_MockReturnValueForever(t *testing.T) {
	f, mocker := NewMockedAddFunc()
	mocker.MockOutputsForever(20, errors.New("error_test"))
	for i := 0; i < 100; i++ {
		sum, err := f(context.Background(), 10, 20)
		assert.NotNil(t, err)
		assert.Equal(t, "error_test", err.Error())
		assert.Equal(t, 20, sum)
	}
}

func TestAddFuncMocker_MockReturnValue(t *testing.T) {
	f, mocker := NewMockedAddFunc()
	mocker.MockOutputs(10, 20, errors.New("error_test"))
	for i := 0; i < 10; i++ {
		sum, err := f(context.Background(), 10, 20)
		assert.NotNil(t, err)
		assert.Equal(t, "error_test", err.Error())
		assert.Equal(t, 20, sum)
	}
}

func TestAddFuncMocker_MockFuncOnce(t *testing.T) {
	f, mocker := NewMockedAddFunc()
	mocker.Mock(1, func(arg1 context.Context, arg2 int, arg3 int) (i int, e error) {
		return arg2 + arg3, nil
	})
	sum, err := f(context.Background(), 10, 20)
	assert.Nil(t, err)
	assert.Equal(t, 30, sum)
}

func TestAddFuncMocker_MockFuncForever(t *testing.T) {
	f, mocker := NewMockedAddFunc()
	mocker.Mock(0, func(arg1 context.Context, arg2 int, arg3 int) (i int, e error) {
		return arg2 + arg3, nil
	})

	for i := 0; i < 100; i++ {
		sum, err := f(context.Background(), 10, 20)
		assert.Nil(t, err)
		assert.Equal(t, 30, sum)
	}
}

func TestAddFuncMocker_MockFunc(t *testing.T) {
	f, mocker := NewMockedAddFunc()
	mocker.Mock(10, func(arg1 context.Context, arg2 int, arg3 int) (i int, e error) {
		return arg2 + arg3, nil
	})

	for i := 0; i < 10; i++ {
		sum, err := f(context.Background(), 10, 20)
		assert.Nil(t, err)
		assert.Equal(t, 30, sum)
	}
}

func TestAddFuncMocker_Invocations(t *testing.T) {
	f, mocker := NewMockedAddFunc()
	mocker.MockOutputsForever(19, nil)
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		f(ctx, 10, 20)
	}

	invocations := mocker.Invocations()
	assert.Equal(t, 10, len(invocations))
	for _, iv := range invocations {
		assert.Equal(t, ctx, iv.Inputs.Ctx)
		assert.Equal(t, 10, iv.Inputs.A)
		assert.Equal(t, 20, iv.Inputs.B)
		assert.Equal(t, 19, iv.Outputs.Sum)
		assert.Nil(t, iv.Outputs.Err)
	}
}

func TestAddFuncMocker_TakeOneInvocation(t *testing.T) {
	f, mocker := NewMockedAddFunc()
	ctx := context.Background()
	err := errors.New("error_test")
	mocker.MockOutputsForever(19, err)
	for i := 0; i < 10; i++ {
		f(ctx, 10, 20)
	}

	for i := 0; i < 10; i++ {
		iv := mocker.TakeOneInvocation()

		assert.Equal(t, ctx, iv.Inputs.Ctx)
		assert.Equal(t, 10, iv.Inputs.A)
		assert.Equal(t, 20, iv.Inputs.B)
		assert.Equal(t, 19, iv.Outputs.Sum)
		assert.Equal(t, err, iv.Outputs.Err)
	}
}
