package tester

import (
	"sync"
)

var LifetimeForever = -1

type Invocation struct {
	Parameters []interface{}
	Returns    []interface{}
}

type CallHandler func(parameters ...interface{}) []interface{}

func NewFuncHandler(f func(parameters ...interface{}) []interface{}) CallHandler {
	return f
}

func NewFixedReturnsFuncHandler(returnValues ...interface{}) CallHandler {
	return func(parameters ...interface{}) []interface{} {
		return returnValues
	}
}

type Mocker struct {
	name string

	handlerLifetime []int
	handlers        []CallHandler
	invocations     []Invocation

	mux *sync.Mutex
}

func NewMocker(name string) *Mocker {
	return &Mocker{
		name: name,

		handlerLifetime: make([]int, 0, 0),
		handlers:        make([]CallHandler, 0, 0),
		invocations:     make([]Invocation, 0, 0),

		mux: &sync.Mutex{},
	}
}

func (f *Mocker) Call(parameters ...interface{}) []interface{} {
	f.mux.Lock()
	defer f.mux.Unlock()

	handler := f.takeOneHandler()
	results := handler(parameters...)

	invocation := Invocation{
		Parameters: parameters,
		Returns:    results,
	}
	f.invocations = append(f.invocations, invocation)

	return results
}

func (f *Mocker) takeOneHandler() (handler CallHandler) {
	if len(f.handlers) == 0 {
		panic(f.name+":no handler found")
	}

	handler = f.handlers[0]

	if f.handlerLifetime[0] == LifetimeForever {
		return
	}

	f.handlerLifetime[0]--
	if f.handlerLifetime[0] == 0 {
		f.handlers = f.handlers[1:]
		f.handlerLifetime = f.handlerLifetime[1:]
	}

	return handler
}

func (f *Mocker) MockOnce(handler CallHandler) *Mocker {
	return f.Mock(1, handler)
}

func (f *Mocker) MockForever(handler CallHandler) *Mocker {
	return f.Mock(LifetimeForever, handler)
}

func (f *Mocker) Mock(nTimes int, handler CallHandler) *Mocker {
	f.mux.Lock()
	defer f.mux.Unlock()

	f.assertValidLifetime(nTimes)
	f.assertLastLifetimeIsNotForever()

	f.handlers = append(f.handlers, handler)
	f.handlerLifetime = append(f.handlerLifetime, nTimes)

	return f
}

func (f *Mocker) assertValidLifetime(nTimes int) {
	if nTimes < 0 && nTimes != LifetimeForever {
		panic(f.name+"the valid lifetime are -1, 0, and positive number")
	}
}

func (f *Mocker) assertLastLifetimeIsNotForever() {
	if len(f.handlerLifetime) == 0 {
		return
	}

	lastLifeTime := f.handlerLifetime[len(f.handlerLifetime)-1]
	if lastLifeTime == LifetimeForever {
		panic(f.name + ":the last handler has been set to forever")
	}
}

func (f *Mocker) Invocations() []Invocation {
	return f.invocations
}

func (f *Mocker) TakeOnceInvocation() Invocation {
	invocation := f.invocations[0]
	f.invocations = f.invocations[1:]
	return invocation
}
