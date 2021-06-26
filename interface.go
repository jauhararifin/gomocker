package gomocker

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/jauhararifin/gotype"
)

type interfaceMockerGenerator struct {
	funcMockerNamer      FuncMockerNamer
	interfaceMockerNamer InterfaceMockerNamer
}

func (f *interfaceMockerGenerator) GenerateInterfaceMocker(name string, interfaceType gotype.InterfaceType) jen.Code {
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
	interfaceType        gotype.InterfaceType
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

	return jen.CustomFunc(jen.Options{}, func(g *jen.Group) {
		g.Add(s.generateInterfaceMockerStruct()).Line().Line()
		g.Add(s.generateMockedInterfaceStruct()).Line().Line()
		g.Add(s.generateMockedInterfaceImpl()).Line().Line()
		g.Add(s.generateInterfaceMockerConstructor()).Line().Line()

		for i := 0; i < len(s.interfaceType.Methods); i++ {
			g.Add(s.funcMockedGenerators[i].generate()).Line().Line()
		}
	})
}

func (s *interfaceMockerGeneratorHelper) generateInterfaceMockerStruct() jen.Code {
	return jen.Type().Id(s.getMockerStructName()).StructFunc(func(g *jen.Group) {
		for _, method := range s.interfaceType.Methods {
			g.Id(method.Name).Op("*").Id(s.funcMockerNamer.MockerName(s.getFuncAlias(method.Name)))
		}
	})
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
	return jen.CustomFunc(jen.Options{}, func(g *jen.Group) {
		for i, method := range s.interfaceType.Methods {
			methodName := method.Name
			g.Func().
				Params(jen.Id("m").Op("*").Id(s.getMockedStructName())).
				Id(methodName).
				Params(s.funcMockedGenerators[i].generateInputParamSignature(true)...).
				Params(s.funcMockedGenerators[i].generateOutputParamSignature(true)...).
				BlockFunc(func(g *jen.Group) {
					if hasOutput := len(method.Func.Outputs) > 0; hasOutput {
						g.Return(jen.Id("m").Dot("mocker").Dot(methodName).Dot("Call").CallFunc(
							s.funcMockedGenerators[i].generateInputListWithEllipsis,
						))
					} else {
						g.Id("m").Dot("mocker").Dot(methodName).Dot("Call").CallFunc(
							s.funcMockedGenerators[i].generateInputListWithEllipsis,
						)
					}
				})
			g.Line()
		}
		g.Line()
	})
}

func (s *interfaceMockerGeneratorHelper) generateInterfaceMockerConstructor() jen.Code {
	return jen.Func().
		Id(s.getConstructorName()).
		Params().
		Params(
			jen.Op("*").Id(s.getMockedStructName()),
			jen.Op("*").Id(s.getMockerStructName()),
		).
		Block(
			jen.Id("m").Op(":=").Op("&").Id(s.getMockerStructName()).ValuesFunc(func(g *jen.Group) {
				for _, method := range s.interfaceType.Methods {
					name := method.Name
					g.Id(name).Op(":").Op("&").Id(s.funcMockerNamer.MockerName(s.getFuncAlias(name))).Values()
				}
			}),
			jen.Return(jen.Op("&").Id(s.getMockedStructName()).Values(jen.Id("m")), jen.Id("m")),
		)
}

func (s *interfaceMockerGeneratorHelper) getConstructorName() string {
	return s.interfaceMockerNamer.ConstructorName(s.interfaceName)
}
