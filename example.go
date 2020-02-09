package tester

import (
	"context"
)

type AddFunc func(ctx context.Context, a int, b int) (int, error)

type AddFuncMocker struct {
	mocker *Mocker
}

type AddFuncMockerInvocation struct {
	Parameter struct {
		p1 context.Context
		p2 int
		p3 int
	}
	Returns struct {
		r1 int
		r2 error
	}
}

func NewMockedAddFunc() (*AddFuncMocker, AddFunc) {
	funcMocker := NewMocker("MockedAddFunc")
	addFuncMocker := &AddFuncMocker{
		mocker: funcMocker,
	}
	return addFuncMocker, func(p1 context.Context, p2 int, p3 int) (r1 int, r2 error) {
		rets := funcMocker.Call(p1, p2, p3)
		r1 = rets[0].(int)
		if rets[1] != nil {
			r2 = rets[1].(error)
		}
		return
	}
}

func (m *AddFuncMocker) MockReturnDefaultValueOnce() {
	var r1 int
	var r2 error
	m.mocker.MockOnce(NewFixedReturnsFuncHandler(r1, r2))
}

func (m *AddFuncMocker) MockReturnDefaultValueForever() {
	var r1 int
	var r2 error
	m.mocker.MockForever(NewFixedReturnsFuncHandler(r1, r2))
}

func (m *AddFuncMocker) MockReturnDefaultValue(nTimes int) {
	var r1 int
	var r2 error
	m.mocker.Mock(nTimes, NewFixedReturnsFuncHandler(r1, r2))
}

func (m *AddFuncMocker) MockReturnValueOnce(r1 int, r2 error) {
	m.mocker.MockOnce(NewFixedReturnsFuncHandler(r1, r2))
}

func (m *AddFuncMocker) MockReturnValueForever(r1 int, r2 error) {
	m.mocker.MockForever(NewFixedReturnsFuncHandler(r1, r2))
}

func (m *AddFuncMocker) MockReturnValue(nTimes int, r1 int, r2 error) {
	m.mocker.Mock(nTimes, NewFixedReturnsFuncHandler(r1, r2))
}

func (m *AddFuncMocker) MockFuncOnce(f func(p1 context.Context, p2 int, p3 int) (int, error)) {
	m.mocker.MockOnce(NewFuncHandler(func(parameters ...interface{}) []interface{} {
		r1, r2 := f(
			parameters[0].(context.Context),
			parameters[1].(int),
			parameters[2].(int),
		)
		return []interface{}{r1, r2}
	}))
}

func (m *AddFuncMocker) MockFuncForever(f func(p1 context.Context, p2 int, p3 int) (int, error)) {
	m.mocker.MockForever(NewFuncHandler(func(parameters ...interface{}) []interface{} {
		r1, r2 := f(
			parameters[0].(context.Context),
			parameters[1].(int),
			parameters[2].(int),
		)
		return []interface{}{r1, r2}
	}))
}

func (m *AddFuncMocker) MockFunc(nTimes int, f func(p1 context.Context, p2 int, p3 int) (int, error)) {
	m.mocker.Mock(nTimes, NewFuncHandler(func(parameters ...interface{}) []interface{} {
		r1, r2 := f(
			parameters[0].(context.Context),
			parameters[1].(int),
			parameters[2].(int),
		)
		return []interface{}{r1, r2}
	}))
}

func (m *AddFuncMocker) Invocations() []AddFuncMockerInvocation {
	invocs := make([]AddFuncMockerInvocation, 0, 0)
	for _, generalInvoc := range m.mocker.invocations {
		invocs = append(invocs, m.convertInvocation(generalInvoc))
	}
	return invocs
}

func (m *AddFuncMocker) TakeOneInvocation() AddFuncMockerInvocation {
	return m.convertInvocation(m.mocker.TakeOnceInvocation())
}

func (m *AddFuncMocker) convertInvocation(invocation Invocation) AddFuncMockerInvocation {
	iv := AddFuncMockerInvocation{}

	iv.Parameter.p1 = invocation.Parameters[0].(context.Context)
	iv.Parameter.p2 = invocation.Parameters[1].(int)
	iv.Parameter.p3 = invocation.Parameters[2].(int)

	iv.Returns.r1 = invocation.Returns[0].(int)

	if invocation.Returns[1] == nil {
		iv.Returns.r2 = nil
	} else {
		iv.Returns.r2 = invocation.Returns[1].(error)
	}

	return iv
}
