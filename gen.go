package gomocker

import (
	"fmt"
	"io"

	"github.com/dave/jennifer/jen"
)

type generateMockerOption struct {
	outputPackagePath string
}

type GenerateMockerOption func(option *generateMockerOption)

func WithOutputPackagePath(outputPackagePath string) GenerateMockerOption {
	return func(option *generateMockerOption) {
		option.outputPackagePath = outputPackagePath
	}
}

type TypeSpec struct {
	PackagePath string
	Name        string
}

type mockerGenerator struct {
	astTypeGenerator interface {
		GenerateTypesFromSpecs(spec ...TypeSpec) []Type
	}
	funcMockerGenerator interface {
		GenerateFunctionMocker(
			name string,
			funcType FuncType,
			withConstructor bool,
		) jen.Code
	}
	interfaceMockerGenerator interface {
		GenerateInterfaceMocker(
			name string,
			interfaceType InterfaceType,
		) jen.Code
	}
}

func (m *mockerGenerator) GenerateMocker(
	specs []TypeSpec,
	w io.Writer,
	options ...GenerateMockerOption,
) error {
	option := m.initOption(options...)

	file := m.createCodeGenFile(option)
	types := m.astTypeGenerator.GenerateTypesFromSpecs(specs...)
	for i, typ := range types {
		file.Add(m.generateEntityMockerByName(option, typ, specs[i].Name)).Line().Line()
	}

	return file.Render(w)
}

func (m *mockerGenerator) initOption(options ...GenerateMockerOption) *generateMockerOption {
	option := &generateMockerOption{}
	for _, opt := range options {
		opt(option)
	}
	return option
}

func (m *mockerGenerator) createCodeGenFile(option *generateMockerOption) *jen.File {
	outputPackagePath := "mock"
	if option.outputPackagePath != "" {
		outputPackagePath = option.outputPackagePath
	}
	file := jen.NewFilePath(outputPackagePath)
	file.HeaderComment("Code generated by gomocker " + gomockerPath + ". DO NOT EDIT.")

	return file
}

func (m *mockerGenerator) generateEntityMockerByName(
	option *generateMockerOption,
	typ Type,
	name string,
) jen.Code {
	if typ.FuncType != nil {
		return m.funcMockerGenerator.GenerateFunctionMocker(name, *typ.FuncType, true)
	}

	if typ.InterfaceType != nil {
		return m.interfaceMockerGenerator.GenerateInterfaceMocker(name, *typ.InterfaceType)
	}

	panic(fmt.Errorf("only supported interface and function"))
}

func (m *mockerGenerator) generateFunctionMocker(
	funcName string,
	funcType FuncType,
	mockerNamer FuncMockerNamer,
) jen.Code {
	funcMockerGenerator := funcMockerGeneratorHelper{
		funcName:        funcName,
		funcType:        funcType,
		mockerNamer:     mockerNamer,
		withConstructor: true,
	}
	return funcMockerGenerator.generate()
}

func (m *mockerGenerator) generateInterfaceMocker(
	interfaceName string,
	interfaceType InterfaceType,
	funcMockerNamer FuncMockerNamer,
	interfaceMockerNamer InterfaceMockerNamer,
) jen.Code {
	generator := &interfaceMockerGeneratorHelper{
		interfaceName:        interfaceName,
		interfaceType:        interfaceType,
		funcMockerNamer:      funcMockerNamer,
		interfaceMockerNamer: interfaceMockerNamer,
	}
	return generator.generate()
}
