// Code generated by gomocker

package examples

import (
	"context"
	gomocker "github.com/jauhararifin/gomocker"
)

type AddFuncMocker struct {
	mocker *gomocker.Mocker
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
	Parameters AddFuncMockerParam
	Returns    AddFuncMockerReturn
}

func NewMockedAddFunc() (*AddFuncMocker, AddFunc) {
	f := gomocker.NewMocker("AddFunc")
	m := &AddFuncMocker{mocker: f}
	return m, func(p1 context.Context, p2 int, p3 int) (r1 int, r2 error) {
		rets := m.parseReturns(f.Call(p1, p2, p3)...)
		return rets.r1, rets.r2
	}
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
func (m *AddFuncMocker) parseReturns(returns ...interface{}) AddFuncMockerReturn {
	r := AddFuncMockerReturn{}
	r.r1 = returns[0].(int)
	if returns[1] != nil {
		r.r2 = returns[1].(error)
	}
	return r
}
func (m *AddFuncMocker) MockReturnDefaultValue(nTimes int) {
	var r1 int
	var r2 error
	m.mocker.Mock(nTimes, gomocker.NewFixedReturnsFuncHandler(r1, r2))
}
func (m *AddFuncMocker) MockReturnDefaultValueForever() {
	m.MockReturnDefaultValue(gomocker.LifetimeForever)
}
func (m *AddFuncMocker) MockReturnDefaultValueOnce() {
	m.MockReturnDefaultValue(1)
}
func (m *AddFuncMocker) MockReturnValues(nTimes int, r1 int, r2 error) {
	m.mocker.Mock(nTimes, gomocker.NewFixedReturnsFuncHandler(r1, r2))
}
func (m *AddFuncMocker) MockReturnValuesForever(r1 int, r2 error) {
	m.MockReturnValues(gomocker.LifetimeForever, r1, r2)
}
func (m *AddFuncMocker) MockReturnValuesOnce(r1 int, r2 error) {
	m.MockReturnValues(1, r1, r2)
}
func (m *AddFuncMocker) Mock(nTimes int, f func(context.Context, int, int) (int, error)) {
	m.mocker.Mock(nTimes, gomocker.NewFuncHandler(func(parameters ...interface{}) []interface{} {
		params := m.parseParams(parameters...)
		r1, r2 := f(params.p1, params.p2, params.p3)
		return []interface{}{r1, r2}
	}))
}
func (m *AddFuncMocker) MockForever(f func(context.Context, int, int) (int, error)) {
	m.Mock(gomocker.LifetimeForever, f)
}
func (m *AddFuncMocker) MockOnce(f func(context.Context, int, int) (int, error)) {
	m.Mock(1, f)
}
func (m *AddFuncMocker) convertInvocation(invocation gomocker.Invocation) AddFuncMockerInvocation {
	iv := AddFuncMockerInvocation{}
	iv.Parameters = m.parseParams(invocation.Parameters...)
	iv.Returns = m.parseReturns(invocation.Returns...)
	return iv
}
func (m *AddFuncMocker) Invocations() []AddFuncMockerInvocation {
	invocs := make([]AddFuncMockerInvocation, 0, 0)
	for _, generalInvoc := range m.mocker.Invocations() {
		invocs = append(invocs, m.convertInvocation(generalInvoc))
	}
	return invocs
}
func (m *AddFuncMocker) TakeOneInvocation() AddFuncMockerInvocation {
	return m.convertInvocation(m.mocker.TakeOneInvocation())
}
