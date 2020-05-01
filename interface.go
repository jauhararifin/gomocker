package gomocker

import (
	"fmt"
	"go/ast"

	"github.com/dave/jennifer/jen"
)

type interfaceMockerGenerator struct {
	interfaceName        string
	interfaceType        *ast.InterfaceType
	funcMockerNamer      Namer
	funcMockedGenerators []*funcMockerGenerator
	exprCodeGen          *exprCodeGenerator
}

func (s *interfaceMockerGenerator) generate() (jen.Code, error) {
	nMethod := s.interfaceType.Methods.NumFields()
	for _, method := range s.interfaceType.Methods.List {
		s.funcMockedGenerators = append(s.funcMockedGenerators, &funcMockerGenerator{
			funcName:        s.getFuncAlias(method.Names[0].String()),
			funcType:        method.Type.(*ast.FuncType),
			mockerNamer:     s.funcMockerNamer,
			withConstructor: false,
			exprCodeGen:     s.exprCodeGen,
		})
	}

	steps := []stepFunc{
		s.generateServiceMockerStruct,
		s.generateMockedInterfaceStruct,
		s.generateMockedInterfaceImpl,
		s.generateInterfaceMockerConstructor,
	}
	for i := 0; i < nMethod; i++ {
		steps = append(steps, s.funcMockedGenerators[i].generate)
	}

	return concatSteps(steps...)
}

func (s *interfaceMockerGenerator) generateServiceMockerStruct() (jen.Code, error) {
	nMethod := s.interfaceType.Methods.NumFields()
	fields := make([]jen.Code, 0, nMethod)
	for _, method := range s.interfaceType.Methods.List {
		methodName := method.Names[0].String()
		fields = append(
			fields,
			jen.Id(methodName).Op("*").Id(s.funcMockerNamer.MockerName(s.getFuncAlias(methodName))),
		)
	}
	return jen.Type().Id(s.getMockerStructName()).Struct(fields...), nil
}

func (s *interfaceMockerGenerator) getFuncAlias(funcName string) string {
	return s.interfaceName + "_" + funcName
}

func (s *interfaceMockerGenerator) getMockerStructName() string {
	// TODO (jauhararifin): create Namer for this
	return s.interfaceName + "Mocker"
}

func (s *interfaceMockerGenerator) generateMockedInterfaceStruct() (jen.Code, error) {
	return jen.Type().Id(s.getMockedStructName()).Struct(jen.Id("mocker").Op("*").Id(s.getMockerStructName())), nil
}

func (s *interfaceMockerGenerator) getMockedStructName() string {
	return "Mocked" + s.interfaceName
}

func (s *interfaceMockerGenerator) generateMockedInterfaceImpl() (jen.Code, error) {
	nMethod := s.interfaceType.Methods.NumFields()
	impls := make([]jen.Code, 0, nMethod)
	for i, method := range s.interfaceType.Methods.List {
		in, out, err := s.funcMockedGenerators[i].generateParamSignature(true)
		if err != nil {
			return nil, err
		}

		_, inputsForCall, _ := s.funcMockedGenerators[i].generateParamList()

		methodName := method.Names[0].String()
		body := jen.Return(jen.Id("m").Dot("mocker").Dot(methodName).Dot("Call").Call(inputsForCall...))
		if method.Type.(*ast.FuncType).Results.NumFields() == 0 {
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
	for _, method := range s.interfaceType.Methods.List {
		name := method.Names[0].String()
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
	return fmt.Sprintf("New%s", s.getMockedStructName())
}
