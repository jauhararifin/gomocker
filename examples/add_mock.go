// Code generated by gomocker github.com/jauhararifin/gomocker. DO NOT EDIT.

package examples

import "sync"

type AddFuncMocker struct {
	mux         sync.Mutex
	handlers    []func(ctx context.Context, a int, b int) (sum int, err error)
	lifetimes   []int
	invocations []struct {
		Inputs struct {
			Ctx context.Context
			A   int
			B   int
		}
		Outputs struct {
			Sum int
			Err error
		}
	}
}

func (m *AddFuncMocker) Mock(nTimes int, f func(ctx context.Context, a int, b int) (sum int, err error)) {
	m.mux.Lock()
	defer m.mux.Unlock()
	nHandler := len(m.lifetimes)
	if nHandler > 0 && m.lifetimes[nHandler-1] == 0 {
		panic("AddFuncMocker: already mocked forever")
	}
	if nTimes < 0 {
		panic("AddFuncMocker: invalid lifetime, valid lifetime are positive number and 0 (0 means forever)")
	}
	m.handlers = append(m.handlers, f)
	m.lifetimes = append(m.lifetimes, nTimes)
}

func (m *AddFuncMocker) MockOnce(f func(ctx context.Context, a int, b int) (sum int, err error)) {
	m.Mock(1, f)
}

func (m *AddFuncMocker) MockForever(f func(ctx context.Context, a int, b int) (sum int, err error)) {
	m.Mock(0, f)
}

func (m *AddFuncMocker) MockOutputs(nTimes int, sum int, err error) {
	m.Mock(nTimes, func(ctx context.Context, a int, b int) (int, error) {
		return sum, err
	})
}

func (m *AddFuncMocker) MockOutputsOnce(sum int, err error) {
	m.MockOutputs(1, sum, err)
}

func (m *AddFuncMocker) MockOutputsForever(sum int, err error) {
	m.MockOutputs(0, sum, err)
}

func (m *AddFuncMocker) MockDefaults(nTimes int) {
	var sum int
	var err error
	m.MockOutputs(nTimes, sum, err)
}

func (m *AddFuncMocker) MockDefaultsOnce() {
	m.MockDefaults(1)
}

func (m *AddFuncMocker) MockDefaultsForever() {
	m.MockDefaults(0)
}

func (m *AddFuncMocker) Call(ctx context.Context, a int, b int) (sum int, err error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if len(m.handlers) == 0 {
		panic("AddFuncMocker: no handler")
	}
	handler := m.handlers[0]
	if m.lifetimes[0] == 1 {
		m.handlers = m.handlers[1:]
		m.lifetimes = m.lifetimes[1:]
	} else if m.lifetimes[0] > 1 {
		m.lifetimes[0]--
	}
	sum, err = handler(ctx, a, b)
	input := struct {
		Ctx context.Context
		A   int
		B   int
	}{ctx, a, b}
	output := struct {
		Sum int
		Err error
	}{sum, err}
	invoc := struct {
		Inputs struct {
			Ctx context.Context
			A   int
			B   int
		}
		Outputs struct {
			Sum int
			Err error
		}
	}{input, output}
	m.invocations = append(m.invocations, invoc)
	return sum, err
}

func (m *AddFuncMocker) Invocations() []struct {
	Inputs struct {
		Ctx context.Context
		A   int
		B   int
	}
	Outputs struct {
		Sum int
		Err error
	}
} {
	return m.invocations
}

func (m *AddFuncMocker) TakeOneInvocation() struct {
	Inputs struct {
		Ctx context.Context
		A   int
		B   int
	}
	Outputs struct {
		Sum int
		Err error
	}
} {
	m.mux.Lock()
	defer m.mux.Unlock()
	if len(m.invocations) == 0 {
		panic("AddFuncMocker: no invocations")
	}
	invoc := m.invocations[0]
	m.invocations = m.invocations[1:]
	return invoc
}

func NewMockedAddFunc() (func(ctx context.Context, a int, b int) (sum int, err error), *AddFuncMocker) {
	m := &AddFuncMocker{}
	return m.Call, m
}
