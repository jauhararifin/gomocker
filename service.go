package gomocker

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/dave/jennifer/jen"
)

type defaultFuncMockedNamer struct{}
func (d defaultFuncMockedNamer) Name(t reflect.Type, i int, public bool) string {
	return t.Method(i).Name + "Mocker"
}

func GenerateServiceMocker(t reflect.Type) (jen.Code, error) {
	if t == nil {
		return nil, errors.New("type is nil")
	}
	if t.Kind() != reflect.Interface {
		return nil, errors.New("type must an interface")
	}

	generator := &serviceMockerGenerator{
		serviceType:          t,
		name:                 t.Name(),
		mockerName:           t.Name() + "Mocker",
		mockedName:           "Mocked" + t.Name(),
		funcMockerNamer:      &defaultFuncMockedNamer{},
		funcMockedGenerators: nil,
	}

	return generator.generate(), nil
}

type serviceMockerGenerator struct {
	serviceType reflect.Type
	name        string
	mockerName  string
	mockedName  string

	funcMockerNamer      Namer
	funcMockedGenerators []*funcMockerGenerator
}

func (s *serviceMockerGenerator) generate() jen.Code {
	nMethod := s.serviceType.NumMethod()
	for i := 0; i < nMethod; i++ {
		m := s.serviceType.Method(i)
		s.funcMockedGenerators = append(s.funcMockedGenerators, &funcMockerGenerator{
			funcType:           m.Type,
			name:               m.Name,
			mockerName:         m.Name + "Mocker",
			inputNamer:         &defaultInputNamer{},
			outputNamer:        &defaultOutputNamer{},
			includeConstructor: false,
		})
	}

	code := jen.Add(s.generateServiceMockerStruct()).Line().
		Add(s.generateMockedServiceStruct()).Line().
		Add(s.generateMockedServiceImpl()...).Line().
		Add(s.generateServiceMockerConstructor()).Line()

	for i := 0; i < nMethod; i++ {
		code = code.Add(s.funcMockedGenerators[i].generate()).Line()
	}

	return code
}

func (s *serviceMockerGenerator) generateServiceMockerStruct() jen.Code {
	nMethod := s.serviceType.NumMethod()
	fields := make([]jen.Code, 0, nMethod)
	for i := 0; i < nMethod; i++ {
		m := s.serviceType.Method(i)
		fields = append(fields, jen.Id(m.Name).Op("*").Id(s.funcMockerNamer.Name(s.serviceType, i, false)))
	}
	return jen.Type().Id(s.mockerName).Struct(fields...)
}

func (s *serviceMockerGenerator) generateMockedServiceStruct() jen.Code {
	return jen.Type().Id(s.mockedName).Struct(jen.Id("mocker").Op("*").Id(s.mockerName))
}

func (s *serviceMockerGenerator) generateMockedServiceImpl() []jen.Code {
	nMethods := s.serviceType.NumMethod()
	impls := make([]jen.Code, 0, 0)
	for i := 0; i < nMethods; i++ {
		m := s.serviceType.Method(i)
		in, out := s.funcMockedGenerators[i].generateParamDef(true)

		_, inputsForCall, _ := s.funcMockedGenerators[i].generateParamList()

		body := jen.Return(jen.Id("m").Dot("mocker").Dot(m.Name).Dot("Call").Call(inputsForCall...))
		if m.Type.NumOut() == 0 {
			body = jen.Id("m").Dot("mocker").Dot(m.Name).Dot("Call").Call(inputsForCall...)
		}

		impls = append(
			impls,
			jen.Func().
				Params(jen.Id("m").Op("*").Id(s.mockedName)).
				Id(m.Name).
				Params(in...).
				Params(out...).
				Block(body).Line(),
		)
	}
	return impls
}

func (s *serviceMockerGenerator) generateServiceMockerConstructor() jen.Code {
	values := make([]jen.Code, 0, 0)
	nMethod := s.serviceType.NumMethod()
	for i := 0; i < nMethod; i++ {
		m := s.serviceType.Method(i)
		values = append(
			values,
			jen.Id(m.Name).Op(":").Op("&").Id(s.funcMockerNamer.Name(s.serviceType, i, true)).Values(),
		)
	}

	return jen.Func().
		Id(fmt.Sprintf("Make%s", s.mockedName)).
		Params().
		Params(
			jen.Op("*").Id(s.mockedName),
			jen.Op("*").Id(s.mockerName),
		).
		Block(
			jen.Id("m").Op(":=").Op("&").Id(s.mockerName).Values(values...),
			jen.Return(jen.Op("&").Id(s.mockedName).Values(jen.Id("m")), jen.Id("m")),
		)
}
