// This code was generated by gomocker

package examples

import (
	"context"
	gomocker "github.com/jauhararifin/gomocker"
	"testing"
)

type Math_Add func(arg1 context.Context, arg2 int, arg3 int) (r1 int, r2 error)
type Math_AddMocker struct {
	mocker *gomocker.Mocker
}
type Math_AddMockerParam struct {
	arg1 context.Context
	arg2 int
	arg3 int
}
type Math_AddMockerReturn struct {
	r1 int
	r2 error
}
type Math_AddMockerInvocation struct {
	Parameters Math_AddMockerParam
	Returns    Math_AddMockerReturn
}

func NewMockedMath_Add(t testing.TB) (*Math_AddMocker, Math_Add) {
	f := gomocker.NewMocker(t, "Math_Add")
	m := &Math_AddMocker{mocker: f}
	return m, m.Call
}
func (m *Math_AddMocker) parseParams(params ...interface{}) Math_AddMockerParam {
	p := Math_AddMockerParam{}
	if params[0] != nil {
		p.arg1 = params[0].(context.Context)
	}
	p.arg2 = params[1].(int)
	p.arg3 = params[2].(int)
	return p
}
func (m *Math_AddMocker) parseReturns(returns ...interface{}) Math_AddMockerReturn {
	r := Math_AddMockerReturn{}
	r.r1 = returns[0].(int)
	if returns[1] != nil {
		r.r2 = returns[1].(error)
	}
	return r
}
func (m *Math_AddMocker) Call(arg1 context.Context, arg2 int, arg3 int) (r1 int, r2 error) {
	rets := m.parseReturns(m.mocker.Call(arg1, arg2, arg3)...)
	return rets.r1, rets.r2
}
func (m *Math_AddMocker) MockReturnDefaultValue(nTimes int) {
	var r1 int
	var r2 error
	m.mocker.Mock(nTimes, gomocker.NewFixedReturnsFuncHandler(r1, r2))
}
func (m *Math_AddMocker) MockReturnDefaultValueForever() {
	m.MockReturnDefaultValue(gomocker.LifetimeForever)
}
func (m *Math_AddMocker) MockReturnDefaultValueOnce() {
	m.MockReturnDefaultValue(1)
}
func (m *Math_AddMocker) MockReturnValues(nTimes int, r1 int, r2 error) {
	m.mocker.Mock(nTimes, gomocker.NewFixedReturnsFuncHandler(r1, r2))
}
func (m *Math_AddMocker) MockReturnValuesForever(r1 int, r2 error) {
	m.MockReturnValues(gomocker.LifetimeForever, r1, r2)
}
func (m *Math_AddMocker) MockReturnValuesOnce(r1 int, r2 error) {
	m.MockReturnValues(1, r1, r2)
}
func (m *Math_AddMocker) Mock(nTimes int, f func(context.Context, int, int) (int, error)) {
	m.mocker.Mock(nTimes, func(parameters ...interface{}) []interface{} {
		params := m.parseParams(parameters...)
		r1, r2 := f(params.arg1, params.arg2, params.arg3)
		return []interface{}{r1, r2}
	})
}
func (m *Math_AddMocker) MockForever(f func(context.Context, int, int) (int, error)) {
	m.Mock(gomocker.LifetimeForever, f)
}
func (m *Math_AddMocker) MockOnce(f func(context.Context, int, int) (int, error)) {
	m.Mock(1, f)
}
func (m *Math_AddMocker) convertInvocation(invocation gomocker.Invocation) Math_AddMockerInvocation {
	iv := Math_AddMockerInvocation{}
	iv.Parameters = m.parseParams(invocation.Parameters...)
	iv.Returns = m.parseReturns(invocation.Returns...)
	return iv
}
func (m *Math_AddMocker) Invocations() []Math_AddMockerInvocation {
	invocs := make([]Math_AddMockerInvocation, 0, 0)
	for _, generalInvoc := range m.mocker.Invocations() {
		invocs = append(invocs, m.convertInvocation(generalInvoc))
	}
	return invocs
}
func (m *Math_AddMocker) TakeOneInvocation() Math_AddMockerInvocation {
	return m.convertInvocation(m.mocker.TakeOneInvocation())
}

type Math_Subtract func(arg1 context.Context, arg2 int, arg3 int) (r1 int, r2 error)
type Math_SubtractMocker struct {
	mocker *gomocker.Mocker
}
type Math_SubtractMockerParam struct {
	arg1 context.Context
	arg2 int
	arg3 int
}
type Math_SubtractMockerReturn struct {
	r1 int
	r2 error
}
type Math_SubtractMockerInvocation struct {
	Parameters Math_SubtractMockerParam
	Returns    Math_SubtractMockerReturn
}

func NewMockedMath_Subtract(t testing.TB) (*Math_SubtractMocker, Math_Subtract) {
	f := gomocker.NewMocker(t, "Math_Subtract")
	m := &Math_SubtractMocker{mocker: f}
	return m, m.Call
}
func (m *Math_SubtractMocker) parseParams(params ...interface{}) Math_SubtractMockerParam {
	p := Math_SubtractMockerParam{}
	if params[0] != nil {
		p.arg1 = params[0].(context.Context)
	}
	p.arg2 = params[1].(int)
	p.arg3 = params[2].(int)
	return p
}
func (m *Math_SubtractMocker) parseReturns(returns ...interface{}) Math_SubtractMockerReturn {
	r := Math_SubtractMockerReturn{}
	r.r1 = returns[0].(int)
	if returns[1] != nil {
		r.r2 = returns[1].(error)
	}
	return r
}
func (m *Math_SubtractMocker) Call(arg1 context.Context, arg2 int, arg3 int) (r1 int, r2 error) {
	rets := m.parseReturns(m.mocker.Call(arg1, arg2, arg3)...)
	return rets.r1, rets.r2
}
func (m *Math_SubtractMocker) MockReturnDefaultValue(nTimes int) {
	var r1 int
	var r2 error
	m.mocker.Mock(nTimes, gomocker.NewFixedReturnsFuncHandler(r1, r2))
}
func (m *Math_SubtractMocker) MockReturnDefaultValueForever() {
	m.MockReturnDefaultValue(gomocker.LifetimeForever)
}
func (m *Math_SubtractMocker) MockReturnDefaultValueOnce() {
	m.MockReturnDefaultValue(1)
}
func (m *Math_SubtractMocker) MockReturnValues(nTimes int, r1 int, r2 error) {
	m.mocker.Mock(nTimes, gomocker.NewFixedReturnsFuncHandler(r1, r2))
}
func (m *Math_SubtractMocker) MockReturnValuesForever(r1 int, r2 error) {
	m.MockReturnValues(gomocker.LifetimeForever, r1, r2)
}
func (m *Math_SubtractMocker) MockReturnValuesOnce(r1 int, r2 error) {
	m.MockReturnValues(1, r1, r2)
}
func (m *Math_SubtractMocker) Mock(nTimes int, f func(context.Context, int, int) (int, error)) {
	m.mocker.Mock(nTimes, func(parameters ...interface{}) []interface{} {
		params := m.parseParams(parameters...)
		r1, r2 := f(params.arg1, params.arg2, params.arg3)
		return []interface{}{r1, r2}
	})
}
func (m *Math_SubtractMocker) MockForever(f func(context.Context, int, int) (int, error)) {
	m.Mock(gomocker.LifetimeForever, f)
}
func (m *Math_SubtractMocker) MockOnce(f func(context.Context, int, int) (int, error)) {
	m.Mock(1, f)
}
func (m *Math_SubtractMocker) convertInvocation(invocation gomocker.Invocation) Math_SubtractMockerInvocation {
	iv := Math_SubtractMockerInvocation{}
	iv.Parameters = m.parseParams(invocation.Parameters...)
	iv.Returns = m.parseReturns(invocation.Returns...)
	return iv
}
func (m *Math_SubtractMocker) Invocations() []Math_SubtractMockerInvocation {
	invocs := make([]Math_SubtractMockerInvocation, 0, 0)
	for _, generalInvoc := range m.mocker.Invocations() {
		invocs = append(invocs, m.convertInvocation(generalInvoc))
	}
	return invocs
}
func (m *Math_SubtractMocker) TakeOneInvocation() Math_SubtractMockerInvocation {
	return m.convertInvocation(m.mocker.TakeOneInvocation())
}

type MathMocker struct {
	Add      *Math_AddMocker
	Subtract *Math_SubtractMocker
}
type MathMocked struct {
	mocker *MathMocker
}

func NewMockedMath(t testing.TB) (*MathMocker, Math) {
	m := &MathMocker{}
	i := &MathMocked{mocker: m}
	return m, i
}
func (m *MathMocked) Add(arg1 context.Context, arg2 int, arg3 int) (r1 int, r2 error) {
	return m.mocker.Add.Call(arg1, arg2, arg3)
}
func (m *MathMocked) Subtract(arg1 context.Context, arg2 int, arg3 int) (r1 int, r2 error) {
	return m.mocker.Subtract.Call(arg1, arg2, arg3)
}
