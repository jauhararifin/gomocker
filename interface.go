package gomocker

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

type interfaceMockerGenerator struct {
	interfaceName        string
	interfaceType        InterfaceType
	interfaceMockerNamer InterfaceMockerNamer
	funcMockerNamer      FuncMockerNamer
	funcMockedGenerators []*funcMockerGenerator
}

func (s *interfaceMockerGenerator) generate() (jen.Code, error) {
	if s.interfaceType.Methods == nil {
		return nil, fmt.Errorf("cannot mock an empty interface")
	}

	for _, method := range s.interfaceType.Methods {
		s.funcMockedGenerators = append(s.funcMockedGenerators, &funcMockerGenerator{
			funcName:        s.getFuncAlias(method.Name),
			funcType:        method.Func,
			mockerNamer:     s.funcMockerNamer,
			withConstructor: false,
		})
	}

	steps := []stepFunc{
		s.generateServiceMockerStruct,
		s.generateMockedInterfaceStruct,
		s.generateMockedInterfaceImpl,
		s.generateInterfaceMockerConstructor,
	}
	for i := 0; i < len(s.interfaceType.Methods); i++ {
		steps = append(steps, s.funcMockedGenerators[i].generate)
	}

	return concatSteps(steps...)
}

func (s *interfaceMockerGenerator) generateServiceMockerStruct() (jen.Code, error) {
	fields := make([]jen.Code, 0, len(s.interfaceType.Methods))
	for _, method := range s.interfaceType.Methods {
		methodName := method.Name
		fields = append(
			fields,
			jen.Id(methodName).Op("*").Id(s.funcMockerNamer.MockerName(s.getFuncAlias(methodName))),
		)
	}
	return jen.Type().Id(s.getMockerStructName()).Struct(fields...), nil
}

func (s *interfaceMockerGenerator) getFuncAlias(funcName string) string {
	return s.interfaceMockerNamer.FunctionAliasName(s.interfaceName, funcName)
}

func (s *interfaceMockerGenerator) getMockerStructName() string {
	return s.interfaceMockerNamer.MockerName(s.interfaceName)
}

func (s *interfaceMockerGenerator) generateMockedInterfaceStruct() (jen.Code, error) {
	return jen.Type().Id(s.getMockedStructName()).Struct(jen.Id("mocker").Op("*").Id(s.getMockerStructName())), nil
}

func (s *interfaceMockerGenerator) getMockedStructName() string {
	return s.interfaceMockerNamer.MockedName(s.interfaceName)
}

func (s *interfaceMockerGenerator) generateMockedInterfaceImpl() (jen.Code, error) {
	impls := make([]jen.Code, 0, len(s.interfaceType.Methods))
	for i, method := range s.interfaceType.Methods {
		in, out := s.funcMockedGenerators[i].generateParamSignature(true, true)
		_, inputsForCall, _ := s.funcMockedGenerators[i].generateParamList()

		methodName := method.Name
		body := jen.Return(jen.Id("m").Dot("mocker").Dot(methodName).Dot("Call").Call(inputsForCall...))
		if len(method.Func.Outputs) == 0 {
			body = jen.Id("m").Dot("mocker").Dot(methodName).Dot("Call").Call(inputsForCall...)
		}

		impls = append(
			impls,
			jen.Func().
				Params(jen.Id("m").Op("*").Id(s.getMockedStructName())).
				Id(methodName).
				Params(in...).
				Params(out...).
				Block(body).Line(),
		)
	}

	j := jen.Empty()
	for _, impl := range impls {
		j.Add(impl)
	}
	return j, nil
}

func (s *interfaceMockerGenerator) generateInterfaceMockerConstructor() (jen.Code, error) {
	values := make([]jen.Code, 0, 0)
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
		), nil
}

func (s *interfaceMockerGenerator) getConstructorName() string {
	return s.interfaceMockerNamer.ConstructorName(s.interfaceName)
}
