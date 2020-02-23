package examples

import (
	"context"
	"sync"
)

type MathMocker struct {
	Add      *AddMocker
	Subtract *SubtractMocker
}
type MockedMath struct {
	mocker *MathMocker
}

func (m *MockedMath) Add(arg1 context.Context, arg2 int, arg3 int) (out1 int, out2 error) {
	return m.mocker.Add.Call(arg1, arg2, arg3)
}
func (m *MockedMath) Subtract(arg1 context.Context, arg2 int, arg3 int) (out1 int, out2 error) {
	return m.mocker.Subtract.Call(arg1, arg2, arg3)
}

func MakeMockedMath() (*MockedMath, *MathMocker) {
	m := &MathMocker{Add: &AddMocker{}, Subtract: &SubtractMocker{}}
	return &MockedMath{m}, m
}

type AddMocker struct {
	mux         sync.Mutex
	handlers    []func(context.Context, int, int) (int, error)
	lifetimes   []int
	invocations []struct {
		Inputs struct {
			Arg1 context.Context
			Arg2 int
			Arg3 int
		}
		Outputs struct {
			Out1 int
			Out2 error
		}
	}
}

func (m *AddMocker) Mock(nTimes int, f func(context.Context, int, int) (int, error)) {
	m.mux.Lock()
	defer m.mux.Unlock()
	nHandler := len(m.lifetimes)
	if nHandler > 0 && m.lifetimes[nHandler-1] == 0 {
		panic("AddMocker: already mocked forever")
	}
	if nTimes < 0 {
		panic("AddMocker: invalid lifetime, valid lifetime are positive number and 0 (0 means forever)")
	}
	m.handlers = append(m.handlers, f)
	m.lifetimes = append(m.lifetimes, nTimes)
}

func (m *AddMocker) MockOnce(f func(context.Context, int, int) (int, error)) {
	m.Mock(1, f)
}

func (m *AddMocker) MockForever(f func(context.Context, int, int) (int, error)) {
	m.Mock(0, f)
}

func (m *AddMocker) MockOutputs(nTimes int, out1 int, out2 error) {
	m.Mock(nTimes, func(arg1 context.Context, arg2 int, arg3 int) (int, error) {
		return out1, out2
	})
}

func (m *AddMocker) MockOutputsOnce(out1 int, out2 error) {
	m.MockOutputs(1, out1, out2)
}

func (m *AddMocker) MockOutputsForever(out1 int, out2 error) {
	m.MockOutputs(0, out1, out2)
}

func (m *AddMocker) MockDefaults(nTimes int) {
	var out1 int
	var out2 error
	m.MockOutputs(nTimes, out1, out2)
}

func (m *AddMocker) MockDefaultsOnce() {
	m.MockDefaults(1)
}

func (m *AddMocker) MockDefaultsForever() {
	m.MockDefaults(0)
}

func (m *AddMocker) Call(arg1 context.Context, arg2 int, arg3 int) (out1 int, out2 error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if len(m.handlers) == 0 {
		panic("AddMocker: no handler")
	}
	handler := m.handlers[0]
	if m.lifetimes[0] == 1 {
		m.handlers = m.handlers[1:]
		m.lifetimes = m.lifetimes[1:]
	} else if m.lifetimes[0] > 1 {
		m.lifetimes[0]--
	}
	out1, out2 = handler(arg1, arg2, arg3)
	input := struct {
		Arg1 context.Context
		Arg2 int
		Arg3 int
	}{arg1, arg2, arg3}
	output := struct {
		Out1 int
		Out2 error
	}{out1, out2}
	invoc := struct {
		Inputs struct {
			Arg1 context.Context
			Arg2 int
			Arg3 int
		}
		Outputs struct {
			Out1 int
			Out2 error
		}
	}{input, output}
	m.invocations = append(m.invocations, invoc)
	return out1, out2
}

func (m *AddMocker) Invocations() []struct {
	Inputs struct {
		Arg1 context.Context
		Arg2 int
		Arg3 int
	}
	Outputs struct {
		Out1 int
		Out2 error
	}
} {
	return m.invocations
}

func (m *AddMocker) TakeOneInvocation() struct {
	Inputs struct {
		Arg1 context.Context
		Arg2 int
		Arg3 int
	}
	Outputs struct {
		Out1 int
		Out2 error
	}
} {
	m.mux.Lock()
	defer m.mux.Unlock()
	if len(m.invocations) == 0 {
		panic("AddMocker: no invocations")
	}
	invoc := m.invocations[0]
	m.invocations = m.invocations[1:]
	return invoc
}

type SubtractMocker struct {
	mux         sync.Mutex
	handlers    []func(context.Context, int, int) (int, error)
	lifetimes   []int
	invocations []struct {
		Inputs struct {
			Arg1 context.Context
			Arg2 int
			Arg3 int
		}
		Outputs struct {
			Out1 int
			Out2 error
		}
	}
}

func (m *SubtractMocker) Mock(nTimes int, f func(context.Context, int, int) (int, error)) {
	m.mux.Lock()
	defer m.mux.Unlock()
	nHandler := len(m.lifetimes)
	if nHandler > 0 && m.lifetimes[nHandler-1] == 0 {
		panic("SubtractMocker: already mocked forever")
	}
	if nTimes < 0 {
		panic("SubtractMocker: invalid lifetime, valid lifetime are positive number and 0 (0 means forever)")
	}
	m.handlers = append(m.handlers, f)
	m.lifetimes = append(m.lifetimes, nTimes)
}

func (m *SubtractMocker) MockOnce(f func(context.Context, int, int) (int, error)) {
	m.Mock(1, f)
}

func (m *SubtractMocker) MockForever(f func(context.Context, int, int) (int, error)) {
	m.Mock(0, f)
}

func (m *SubtractMocker) MockOutputs(nTimes int, out1 int, out2 error) {
	m.Mock(nTimes, func(arg1 context.Context, arg2 int, arg3 int) (int, error) {
		return out1, out2
	})
}

func (m *SubtractMocker) MockOutputsOnce(out1 int, out2 error) {
	m.MockOutputs(1, out1, out2)
}

func (m *SubtractMocker) MockOutputsForever(out1 int, out2 error) {
	m.MockOutputs(0, out1, out2)
}

func (m *SubtractMocker) MockDefaults(nTimes int) {
	var out1 int
	var out2 error
	m.MockOutputs(nTimes, out1, out2)
}

func (m *SubtractMocker) MockDefaultsOnce() {
	m.MockDefaults(1)
}

func (m *SubtractMocker) MockDefaultsForever() {
	m.MockDefaults(0)
}

func (m *SubtractMocker) Call(arg1 context.Context, arg2 int, arg3 int) (out1 int, out2 error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if len(m.handlers) == 0 {
		panic("SubtractMocker: no handler")
	}
	handler := m.handlers[0]
	if m.lifetimes[0] == 1 {
		m.handlers = m.handlers[1:]
		m.lifetimes = m.lifetimes[1:]
	} else if m.lifetimes[0] > 1 {
		m.lifetimes[0]--
	}
	out1, out2 = handler(arg1, arg2, arg3)
	input := struct {
		Arg1 context.Context
		Arg2 int
		Arg3 int
	}{arg1, arg2, arg3}
	output := struct {
		Out1 int
		Out2 error
	}{out1, out2}
	invoc := struct {
		Inputs struct {
			Arg1 context.Context
			Arg2 int
			Arg3 int
		}
		Outputs struct {
			Out1 int
			Out2 error
		}
	}{input, output}
	m.invocations = append(m.invocations, invoc)
	return out1, out2
}

func (m *SubtractMocker) Invocations() []struct {
	Inputs struct {
		Arg1 context.Context
		Arg2 int
		Arg3 int
	}
	Outputs struct {
		Out1 int
		Out2 error
	}
} {
	return m.invocations
}

func (m *SubtractMocker) TakeOneInvocation() struct {
	Inputs struct {
		Arg1 context.Context
		Arg2 int
		Arg3 int
	}
	Outputs struct {
		Out1 int
		Out2 error
	}
} {
	m.mux.Lock()
	defer m.mux.Unlock()
	if len(m.invocations) == 0 {
		panic("SubtractMocker: no invocations")
	}
	invoc := m.invocations[0]
	m.invocations = m.invocations[1:]
	return invoc
}
