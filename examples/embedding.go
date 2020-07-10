package examples

import (
	"fmt"
	"io"
)

type InterfaceA interface {
	A()
}

type InterfaceB interface {
	B()
}

type InterfaceC interface {
	InterfaceA
	InterfaceB
	fmt.Stringer
	io.ReadWriteCloser
	C()
}