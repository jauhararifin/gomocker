package gomocker

import (
	"context"
)

type AddFunc func(ctx context.Context, a int, b int) (int, error)

type AddFuncMocker struct {
	mocker *Mocker
}

type AddFuncMockerParam struct {
	p1 context.Context
	p2 int
	p3 int
}

type AddFuncMockerReturn struct {
	r1 int
	r2 error
}

type AddFuncMockerInvocation struct {
	Parameter AddFuncMockerParam
	Returns   AddFuncMockerReturn
}

func NewMockedAddFunc() (*AddFuncMocker, AddFunc) {
	funcMocker := NewMocker("MockedAddFunc")
	addFuncMocker := &AddFuncMocker{
		mocker: funcMocker,
	}
	return addFuncMocker, func(p1 context.Context, p2 int, p3 int) (r1 int, r2 error) {
		rets := addFuncMocker.parseReturns(funcMocker.Call(p1, p2, p3)...)
		return rets.r1, rets.r2
	}
}

func (m *AddFuncMocker) parseReturns(returns ...interface{}) AddFuncMockerReturn {
	r := AddFuncMockerReturn{}
	r.r1 = returns[0].(int)
	if returns[1] != nil {
		r.r2 = returns[1].(error)
	}
	return r
}

func (m *AddFuncMocker) parseParams(params ...interface{}) AddFuncMockerParam {
	p := AddFuncMockerParam{}
	if params[0] != nil {
		p.p1 = params[0].(context.Context)
	}
	p.p2 = params[1].(int)
	p.p3 = params[2].(int)
	return p
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
		params := m.parseParams(parameters...)
		r1, r2 := f(params.p1, params.p2, params.p3)
		return []interface{}{r1, r2}
	}))
}

func (m *AddFuncMocker) MockFuncForever(f func(p1 context.Context, p2 int, p3 int) (int, error)) {
	m.mocker.MockForever(NewFuncHandler(func(parameters ...interface{}) []interface{} {
		params := m.parseParams(parameters...)
		r1, r2 := f(params.p1, params.p2, params.p3)
		return []interface{}{r1, r2}
	}))
}

func (m *AddFuncMocker) MockFunc(nTimes int, f func(p1 context.Context, p2 int, p3 int) (int, error)) {
	m.mocker.Mock(nTimes, NewFuncHandler(func(parameters ...interface{}) []interface{} {
		params := m.parseParams(parameters...)
		r1, r2 := f(params.p1, params.p2, params.p3)
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
	return m.convertInvocation(m.mocker.TakeOneInvocation())
}

func (m *AddFuncMocker) convertInvocation(invocation Invocation) AddFuncMockerInvocation {
	iv := AddFuncMockerInvocation{}
	iv.Parameter = m.parseParams(invocation.Parameters...)
	iv.Returns = m.parseReturns(invocation.Returns...)
	return iv
}
