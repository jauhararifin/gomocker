package examples

import "sync"

type AggregatorMocker struct {
	SumInt *SumIntMocker
}
type MockedAggregator struct {
	mocker *AggregatorMocker
}

func (m *MockedAggregator) SumInt(arg1 ...int) (out1 int) {
	return m.mocker.SumInt.Call(arg1...)
}

func MakeMockedAggregator() (*MockedAggregator, *AggregatorMocker) {
	m := &AggregatorMocker{SumInt: &SumIntMocker{}}
	return &MockedAggregator{m}, m
}

type SumIntMocker struct {
	mux         sync.Mutex
	handlers    []func(...int) int
	lifetimes   []int
	invocations []struct {
		Inputs struct {
			Arg1 []int
		}
		Outputs struct {
			Out1 int
		}
	}
}

func (m *SumIntMocker) Mock(nTimes int, f func(...int) int) {
	m.mux.Lock()
	defer m.mux.Unlock()
	nHandler := len(m.lifetimes)
	if nHandler > 0 && m.lifetimes[nHandler-1] == 0 {
		panic("SumIntMocker: already mocked forever")
	}
	if nTimes < 0 {
		panic("SumIntMocker: invalid lifetime, valid lifetime are positive number and 0 (0 means forever)")
	}
	m.handlers = append(m.handlers, f)
	m.lifetimes = append(m.lifetimes, nTimes)
}

func (m *SumIntMocker) MockOnce(f func(...int) int) {
	m.Mock(1, f)
}

func (m *SumIntMocker) MockForever(f func(...int) int) {
	m.Mock(0, f)
}

func (m *SumIntMocker) MockOutputs(nTimes int, out1 int) {
	m.Mock(nTimes, func(arg1 ...int) int {
		return out1
	})
}

func (m *SumIntMocker) MockOutputsOnce(out1 int) {
	m.MockOutputs(1, out1)
}

func (m *SumIntMocker) MockOutputsForever(out1 int) {
	m.MockOutputs(0, out1)
}

func (m *SumIntMocker) MockDefaults(nTimes int) {
	var out1 int
	m.MockOutputs(nTimes, out1)
}

func (m *SumIntMocker) MockDefaultsOnce() {
	m.MockDefaults(1)
}

func (m *SumIntMocker) MockDefaultsForever() {
	m.MockDefaults(0)
}

func (m *SumIntMocker) Call(arg1 ...int) (out1 int) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if len(m.handlers) == 0 {
		panic("SumIntMocker: no handler")
	}
	handler := m.handlers[0]
	if m.lifetimes[0] == 1 {
		m.handlers = m.handlers[1:]
		m.lifetimes = m.lifetimes[1:]
	} else if m.lifetimes[0] > 1 {
		m.lifetimes[0]--
	}
	out1 = handler(arg1...)
	input := struct {
		Arg1 []int
	}{arg1}
	output := struct {
		Out1 int
	}{out1}
	invoc := struct {
		Inputs struct {
			Arg1 []int
		}
		Outputs struct {
			Out1 int
		}
	}{input, output}
	m.invocations = append(m.invocations, invoc)
	return out1
}

func (m *SumIntMocker) Invocations() []struct {
	Inputs struct {
		Arg1 []int
	}
	Outputs struct {
		Out1 int
	}
} {
	return m.invocations
}

func (m *SumIntMocker) TakeOneInvocation() struct {
	Inputs struct {
		Arg1 []int
	}
	Outputs struct {
		Out1 int
	}
} {
	m.mux.Lock()
	defer m.mux.Unlock()
	if len(m.invocations) == 0 {
		panic("SumIntMocker: no invocations")
	}
	invoc := m.invocations[0]
	m.invocations = m.invocations[1:]
	return invoc
}
