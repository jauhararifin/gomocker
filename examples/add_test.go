package examples

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddFuncMocker_MockReturnDefaultValueOnce(t *testing.T) {
	mocker, f := NewMockedAddFunc()
	mocker.MockReturnDefaultValueOnce()
	sum, err := f(context.Background(), 10, 20)
	assert.Nil(t, err, "unexpected error when calling AddFunc")
	assert.Equal(t, 0, sum, "sum should contain default value, which is zero")

	defer func() {
		assert.NotNil(t, recover())
	}()
	f(context.Background(), 10, 20)
}

func TestAddFuncMocker_MockReturnDefaultValueForever(t *testing.T) {
	mocker, f := NewMockedAddFunc()
	mocker.MockReturnDefaultValueForever()
	for i := 0; i < 100; i++ {
		sum, err := f(context.Background(), 10, 20)
		assert.Nil(t, err, "unexpected error when calling AddFunc")
		assert.Equal(t, 0, sum, "sum should contain default value, which is zero")
	}
}

func TestAddFuncMocker_MockReturnDefaultValue(t *testing.T) {
	mocker, f := NewMockedAddFunc()
	mocker.MockReturnDefaultValue(10)
	for i := 0; i < 10; i++ {
		sum, err := f(context.Background(), 10, 20)
		assert.Nil(t, err, "unexpected error when calling AddFunc")
		assert.Equal(t, 0, sum, "sum should contain default value, which is zero")
	}

	defer func() {
		assert.NotNil(t, recover())
	}()
	f(context.Background(), 10, 20)
}

func TestAddFuncMocker_MockReturnValueOnce(t *testing.T) {
	mocker, f := NewMockedAddFunc()
	mocker.MockReturnValuesOnce(20, errors.New("error_test"))
	sum, err := f(context.Background(), 10, 20)
	assert.NotNil(t, err)
	assert.Equal(t, "error_test", err.Error())
	assert.Equal(t, 20, sum)

	defer func() {
		assert.NotNil(t, recover())
	}()
	f(context.Background(), 10, 20)
}

func TestAddFuncMocker_MockReturnValueForever(t *testing.T) {
	mocker, f := NewMockedAddFunc()
	mocker.MockReturnValuesForever(20, errors.New("error_test"))
	for i := 0; i < 100; i++ {
		sum, err := f(context.Background(), 10, 20)
		assert.NotNil(t, err)
		assert.Equal(t, "error_test", err.Error())
		assert.Equal(t, 20, sum)
	}
}

func TestAddFuncMocker_MockReturnValue(t *testing.T) {
	mocker, f := NewMockedAddFunc()
	mocker.MockReturnValues(10, 20, errors.New("error_test"))
	for i := 0; i < 10; i++ {
		sum, err := f(context.Background(), 10, 20)
		assert.NotNil(t, err)
		assert.Equal(t, "error_test", err.Error())
		assert.Equal(t, 20, sum)
	}

	defer func() {
		assert.NotNil(t, recover())
	}()
	f(context.Background(), 10, 20)
}

func TestAddFuncMocker_MockFuncOnce(t *testing.T) {
	mocker, f := NewMockedAddFunc()
	mocker.MockOnce(func(p1 context.Context, p2 int, p3 int) (i int, e error) {
		return p2 + p3, nil
	})
	sum, err := f(context.Background(), 10, 20)
	assert.Nil(t, err)
	assert.Equal(t, 30, sum)

	defer func() {
		assert.NotNil(t, recover())
	}()
	f(context.Background(), 10, 20)
}

func TestAddFuncMocker_MockFuncForever(t *testing.T) {
	mocker, f := NewMockedAddFunc()
	mocker.MockForever(func(p1 context.Context, p2 int, p3 int) (i int, e error) {
		return p2 + p3, nil
	})

	for i := 0; i < 100; i++ {
		sum, err := f(context.Background(), 10, 20)
		assert.Nil(t, err)
		assert.Equal(t, 30, sum)
	}
}

func TestAddFuncMocker_MockFunc(t *testing.T) {
	mocker, f := NewMockedAddFunc()
	mocker.Mock(10, func(p1 context.Context, p2 int, p3 int) (i int, e error) {
		return p2 + p3, nil
	})

	for i := 0; i < 10; i++ {
		sum, err := f(context.Background(), 10, 20)
		assert.Nil(t, err)
		assert.Equal(t, 30, sum)
	}

	defer func() {
		assert.NotNil(t, recover())
	}()
	f(context.Background(), 10, 20)
}

func TestAddFuncMocker_Invocations(t *testing.T) {
	mocker, f := NewMockedAddFunc()
	mocker.MockReturnValuesForever(19, nil)
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		f(ctx, 10, 20)
	}

	invocations := mocker.Invocations()
	assert.Equal(t, 10, len(invocations))
	for _, iv := range invocations {
		assert.Equal(t, ctx, iv.Parameters.p1)
		assert.Equal(t, 10, iv.Parameters.p2)
		assert.Equal(t, 20, iv.Parameters.p3)
		assert.Equal(t, 19, iv.Returns.r1)
		assert.Nil(t, iv.Returns.r2)
	}
}

func TestAddFuncMocker_TakeOneInvocation(t *testing.T) {
	mocker, f := NewMockedAddFunc()
	ctx := context.Background()
	err := errors.New("error_test")
	mocker.MockReturnValuesForever(19, err)
	for i := 0; i < 10; i++ {
		f(ctx, 10, 20)
	}

	for i := 0; i < 10; i++ {
		iv := mocker.TakeOneInvocation()

		assert.Equal(t, ctx, iv.Parameters.p1)
		assert.Equal(t, 10, iv.Parameters.p2)
		assert.Equal(t, 20, iv.Parameters.p3)
		assert.Equal(t, 19, iv.Returns.r1)
		assert.Equal(t, err, iv.Returns.r2)
	}

	defer func() {
		assert.NotNil(t, recover())
	}()
	mocker.TakeOneInvocation()
}

