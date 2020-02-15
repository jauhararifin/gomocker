package gomocker

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type SimpleAdd func(ctx context.Context, a, b int) (int, error)

type SimpleAddParams struct {
	Ctx context.Context
	A   int
	B   int
}

type SimpleAddReturns struct {
	Sum int
	Err error
}

type SimpleAddInvocation struct {
	Parameters SimpleAddParams
	Returns    SimpleAddReturns
}

func TestReflectMocker_MockReturnDefaultValues(t *testing.T) {
	m := NewReflectMocker(t, "SimpleAdd", SimpleAddInvocation{})

	t.Run("normal flow", func(t *testing.T) {
		m.MockReturnDefaultValues(1)
		r := m.Call(context.Background(), 1, 2).(SimpleAddReturns)
		assert.Equal(t, 0, r.Sum)
		assert.Equal(t, nil, r.Err)
	})

	t.Run("nil input", func(t *testing.T) {
		m.MockReturnDefaultValues(1)
		r := m.Call(nil, 1, 2).(SimpleAddReturns)
		assert.Equal(t, 0, r.Sum)
		assert.Equal(t, nil, r.Err)
	})
}

func TestReflectMocker_MockReturnValues(t *testing.T) {
	m := NewReflectMocker(t, "SimpleAdd", SimpleAddInvocation{})
	dummyErr := errors.New("dummy_error")
	m.MockReturnValues(1, 100, dummyErr)
	r := m.Call(context.Background(), 1, 2).(SimpleAddReturns)
	assert.Equal(t, 100, r.Sum)
	assert.True(t, errors.Is(r.Err, dummyErr))
}

func TestReflectMocker_Mock(t *testing.T) {
	m := NewReflectMocker(t, "SimpleAdd", SimpleAddInvocation{})
	m.Mock(1, func(ctx context.Context, a, b int) (int, error) {
		assert.Equal(t, context.Background(), ctx)
		assert.Equal(t, 1, a)
		assert.Equal(t, 2, b)
		return 101, nil
	})
	r := m.Call(context.Background(), 1, 2).(SimpleAddReturns)
	assert.Equal(t, 101, r.Sum)
	assert.Nil(t, r.Err)
}

func TestReflectMocker_TakeOneInvocation(t *testing.T) {
	m := NewReflectMocker(t, "SimpleAdd", SimpleAddInvocation{})
	m.MockForever(func(ctx context.Context, a, b int) (int, error) {
		return a + b, nil
	})

	for i := 0; i < 100; i++ {
		r := m.Call(context.Background(), i, i+1).(SimpleAddReturns)
		assert.Equal(t, i+i+1, r.Sum)
		assert.Nil(t, r.Err)
	}

	for i := 0; i < 100; i++ {
		invoc := m.TakeOneInvocation().(SimpleAddInvocation)
		assert.Equal(t, context.Background(), invoc.Parameters.Ctx)
		assert.Equal(t, i, invoc.Parameters.A)
		assert.Equal(t, i+1, invoc.Parameters.B)
		assert.Equal(t, i+i+1, invoc.Returns.Sum)
		assert.Nil(t, invoc.Returns.Err)
	}
}

func TestReflectMocker_Invocations(t *testing.T) {
	m := NewReflectMocker(t, "SimpleAdd", SimpleAddInvocation{})
	m.MockForever(func(ctx context.Context, a, b int) (int, error) {
		return a + b, nil
	})

	for i := 0; i < 100; i++ {
		r := m.Call(context.Background(), i, i+1).(SimpleAddReturns)
		assert.Equal(t, i+i+1, r.Sum)
		assert.Nil(t, r.Err)
	}

	invocsInterface := m.Invocations()
	invocs := make([]SimpleAddInvocation, len(invocsInterface), len(invocsInterface))
	for i, iv := range invocsInterface {
		invocs[i] = iv.(SimpleAddInvocation)
	}

	assert.Equal(t, 100, len(invocs))
	for i := 0; i < 100; i++ {
		invoc := invocs[i]
		assert.Equal(t, context.Background(), invoc.Parameters.Ctx)
		assert.Equal(t, i, invoc.Parameters.A)
		assert.Equal(t, i+1, invoc.Parameters.B)
		assert.Equal(t, i+i+1, invoc.Returns.Sum)
		assert.Nil(t, invoc.Returns.Err)
	}
}
