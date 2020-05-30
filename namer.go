package gomocker

import (
	"fmt"
	"strings"
)

type FuncMockerNamer interface {
	MockerName(identifier string) string
	ConstructorName(identifier string) string
	InvocationName(identifier string) string
	ArgumentName(i int, originalArgName string, needPublic bool, isReturnArgument bool) string
}

type InterfaceMockerNamer interface {
	MockerName(identifier string) string
	MockedName(identifier string) string
	ConstructorName(identifier string) string
	FunctionAliasName(identifier, functionName string) string
}

type defaultFuncMockerNamer struct{}

func (*defaultFuncMockerNamer) MockerName(typeName string) string {
	return makePublic(typeName) + "Mocker"
}

func (*defaultFuncMockerNamer) ConstructorName(typeName string) string {
	return "NewMocked" + makePublic(typeName)
}

func (*defaultFuncMockerNamer) InvocationName(typeName string) string {
	return makePublic(typeName) + "Invocation"
}

func (*defaultFuncMockerNamer) ArgumentName(i int, name string, needPublic bool, isReturnArgument bool) string {
	if name != "" {
		if needPublic {
			return strings.ToUpper(name[:1]) + name[1:]
		}
		return name
	} else if isReturnArgument {
		if needPublic {
			return fmt.Sprintf("Out%d", i+1)
		}
		return fmt.Sprintf("out%d", i+1)
	} else if needPublic {
		return fmt.Sprintf("Arg%d", i+1)
	}
	return fmt.Sprintf("arg%d", i+1)
}

type defaultInterfaceMockerNamer struct{}

func (d *defaultInterfaceMockerNamer) MockerName(identifier string) string {
	return makePublic(identifier) + "Mocker"
}

func (d *defaultInterfaceMockerNamer) MockedName(identifier string) string {
	return "Mocked" + makePublic(identifier)
}

func (d *defaultInterfaceMockerNamer) ConstructorName(identifier string) string {
	return "NewMocked" + makePublic(identifier)
}

func (d *defaultInterfaceMockerNamer) FunctionAliasName(identifier, functionName string) string {
	return makePublic(identifier) + "_" + makePublic(functionName)
}
