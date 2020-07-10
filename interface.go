package gomocker

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

type interfaceMockerGenerator struct {
	funcMockerNamer      FuncMockerNamer
	interfaceMockerNamer InterfaceMockerNamer
}

func (f *interfaceMockerGenerator) GenerateInterfaceMocker(name string, interfaceType InterfaceType) jen.Code {
	generator := &interfaceMockerGeneratorHelper{
		interfaceName:        name,
		interfaceType:        interfaceType,
		funcMockerNamer:      f.funcMockerNamer,
		interfaceMockerNamer: f.interfaceMockerNamer,
	}
	return generator.generate()
}

type interfaceMockerGeneratorHelper struct {
	interfaceName        string
	interfaceType        InterfaceType
	interfaceMockerNamer InterfaceMockerNamer
	funcMockerNamer      FuncMockerNamer
	funcMockedGenerators []*funcMockerGeneratorHelper
}

func (s *interfaceMockerGeneratorHelper) generate() jen.Code {
	if s.interfaceType.Methods == nil {
		panic(fmt.Errorf("cannot mock an empty interface"))
	}

	for _, method := range s.interfaceType.Methods {
		// TODO (jauhararifin): maybe can make the funcMockerGeneratorHelper more modular
		s.funcMockedGenerators = append(s.funcMockedGenerators, &funcMockerGeneratorHelper{
			funcName:        s.getFuncAlias(method.Name),
			funcType:        method.Func,
			mockerNamer:     s.funcMockerNamer,
			withConstructor: false,
		})
	}

	steps := []stepFunc{
		s.generateInterfaceMockerStruct,
		s.generateMockedInterfaceStruct,
		s.generateMockedInterfaceImpl,
		s.generateInterfaceMockerConstructor,
	}
	for i := 0; i < len(s.interfaceType.Methods); i++ {
		steps = append(steps, s.funcMockedGenerators[i].generate)
	}

	return concatSteps(steps...)
}

func (s *interfaceMockerGeneratorHelper) generateInterfaceMockerStruct() jen.Code {
	fields := make([]jen.Code, 0, len(s.interfaceType.Methods))
	for _, method := range s.interfaceType.Methods {
		methodName := method.Name
		fields = append(
			fields,
			jen.Id(methodName).Op("*").Id(s.funcMockerNamer.MockerName(s.getFuncAlias(methodName))),
		)
	}
	return jen.Type().Id(s.getMockerStructName()).Struct(fields...)
}

func (s *interfaceMockerGeneratorHelper) getFuncAlias(funcName string) string {
	return s.interfaceMockerNamer.FunctionAliasName(s.interfaceName, funcName)
}

func (s *interfaceMockerGeneratorHelper) getMockerStructName() string {
	return s.interfaceMockerNamer.MockerName(s.interfaceName)
}

func (s *interfaceMockerGeneratorHelper) generateMockedInterfaceStruct() jen.Code {
	return jen.Type().Id(s.getMockedStructName()).Struct(jen.Id("mocker").Op("*").Id(s.getMockerStructName()))
}

func (s *interfaceMockerGeneratorHelper) getMockedStructName() string {
	return s.interfaceMockerNamer.MockedName(s.interfaceName)
}

func (s *interfaceMockerGeneratorHelper) generateMockedInterfaceImpl() jen.Code {
	code := jen.Empty()

	for i, method := range s.interfaceType.Methods {
		methodName := method.Name
		impl := jen.Func().
			Params(jen.Id("m").Op("*").Id(s.getMockedStructName())).
			Id(methodName).
			Params(s.funcMockedGenerators[i].generateInputParamSignature(true)...).
			Params(s.funcMockedGenerators[i].generateOutputParamSignature(true)...)

		if hasOutput := len(method.Func.Outputs) > 0; hasOutput {
			impl.Block(
				jen.Return(jen.Id("m").Dot("mocker").Dot(methodName).Dot("Call").Call(
					s.funcMockedGenerators[i].generateInputList(true)...,
				)),
			)
		} else {
			impl.Block(
				jen.Id("m").Dot("mocker").Dot(methodName).Dot("Call").Call(
					s.funcMockedGenerators[i].generateInputList(true)...,
				),
			)
		}

		code.Add(impl).Line()
	}

	return code.Line()
}

func (s *interfaceMockerGeneratorHelper) generateInterfaceMockerConstructor() jen.Code {
	values := make([]jen.Code, 0)
	for _, method := range s.interfaceType.Methods {
		name := method.Name
		values = append(
			values,
			jen.Id(name).Op(":").Op("&").Id(s.funcMockerNamer.MockerName(s.getFuncAlias(name))).Values(),
		)
	}

	return jen.Func().
		Id(s.getConstructorName()).
		Params().
		Params(
			jen.Op("*").Id(s.getMockedStructName()),
			jen.Op("*").Id(s.getMockerStructName()),
		).
		Block(
			jen.Id("m").Op(":=").Op("&").Id(s.getMockerStructName()).Values(values...),
			jen.Return(jen.Op("&").Id(s.getMockedStructName()).Values(jen.Id("m")), jen.Id("m")),
		)
}

func (s *interfaceMockerGeneratorHelper) getConstructorName() string {
	return s.interfaceMockerNamer.ConstructorName(s.interfaceName)
}
