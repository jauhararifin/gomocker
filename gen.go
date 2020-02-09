package gomocker

import (
	"errors"
	"fmt"
	"github.com/dave/jennifer/jen"
	"io"
	"reflect"
)

func GenerateMockedFunction(function interface{}, packageName string, w io.Writer) error {
	ftype := reflect.TypeOf(function)
	if ftype.Kind() != reflect.Func {
		return errors.New("not_a_function")
	}
	m := &mockedFunctionGenerator{ftype: ftype}
	return m.Generate(packageName, w)
}

type mockedFunctionGenerator struct {
	ftype reflect.Type
}

func (m *mockedFunctionGenerator) Generate(packageName string, w io.Writer) error {
	code := jen.NewFile(packageName)

	code.
		Add(m.generateMockerStruct()).Line().
		Add(m.generateParamStruct()).Line().
		Add(m.generateReturnStruct()).Line().
		Add(m.generateConstructor()).Line().
		Add(m.generateParseParams()).Line().
		Add(m.generateParseReturns()).Line().
		Add(m.generateMockReturnDefaultValue()).Line().
		Add(m.generateMockReturnDefaultValueForever()).Line().
		Add(m.generateMockReturnDefaultValueOnce()).Line().
		Add(m.generateMockReturnValues()).Line().
		Add(m.generateMockReturnValuesForever()).Line().
		Add(m.generateMockReturnValuesOnce()).Line().
		Add(m.generateMock()).Line().
		Add(m.generateMockForever()).Line().
		Add(m.generateMockOnce()).Line()

	return code.Render(w)
}

func (m *mockedFunctionGenerator) generateMockerStruct() jen.Code {
	return jen.Type().Id(m.mockerStructName()).Struct(jen.Id("mocker").Op("*").Qual(m.gomockerPackage(), "Mocker"))
}

func (m *mockedFunctionGenerator) mockerStructName() string {
	return m.functionName() + "Mocker"
}

func (m *mockedFunctionGenerator) functionName() string {
	return m.ftype.Name()
}

func (m *mockedFunctionGenerator) gomockerPackage() string {
	return "github.com/jauhararifin/gomocker"
}

func (m *mockedFunctionGenerator) generateParamStruct() jen.Code {
	return jen.Type().Id(m.mockerParamStructName()).Struct(m.generateInputParamDefinitions()...)
}

func (m *mockedFunctionGenerator) mockerParamStructName() string {
	return m.functionName() + "MockerParam"
}

func (m *mockedFunctionGenerator) generateInputParamDefinitions() []jen.Code {
	params := make([]jen.Code, m.ftype.NumIn(), m.ftype.NumIn())
	for i := 0; i < m.ftype.NumIn(); i++ {
		name := fmt.Sprintf("p%d", i+1)
		params[i] = jen.Id(name).Add(generateJenFromType(m.ftype.In(i)))
	}
	return params
}

func (m *mockedFunctionGenerator) generateReturnStruct() jen.Code {
	return jen.Type().Id(m.mockerReturnStructName()).Struct(m.generateOutputParamDefinitions()...)
}

func (m *mockedFunctionGenerator) mockerReturnStructName() string {
	return m.functionName() + "MockerReturn"
}

func (m *mockedFunctionGenerator) generateOutputParamDefinitions() []jen.Code {
	params := make([]jen.Code, m.ftype.NumOut(), m.ftype.NumOut())
	for i := 0; i < m.ftype.NumOut(); i++ {
		name := fmt.Sprintf("r%d", i+1)
		params[i] = jen.Id(name).Add(generateJenFromType(m.ftype.Out(i)))
	}
	return params
}

func (m *mockedFunctionGenerator) generateInvocationStruct() jen.Code {
	return jen.Type().Id(m.mockerInvocationStructName()).Struct(
		jen.Id(m.invocationParametersFieldName()).Id(m.mockerParamStructName()),
		jen.Id(m.invocationReturnsFieldName()).Id(m.mockerReturnStructName()),
	)
}

func (m *mockedFunctionGenerator) mockerInvocationStructName() string {
	return m.functionName() + "MockerInvocation"
}

func (m *mockedFunctionGenerator) invocationParametersFieldName() string {
	return "Parameters"
}

func (m *mockedFunctionGenerator) invocationReturnsFieldName() string {
	return "Returns"
}

func (m *mockedFunctionGenerator) generateConstructor() jen.Code {
	return jen.Func().Id(m.mockerConstructorName()).
		Params().
		Params(jen.Op("*").Id(m.mockerStructName()), jen.Id(m.functionName())).
		Block(
			jen.Id("f").Op(":=").Qual(m.gomockerPackage(), "NewMocker").Call(jen.Lit(m.functionName())),
			jen.Id("m").Op(":=").Op("&").Id(m.mockerStructName()).Values(
				jen.Id("mocker").Op(":").Id("f"),
			),
			jen.Return(
				jen.Id("m"),
				jen.Func().
					Params(m.generateInputParamDefinitions()...).
					Params(m.generateOutputParamDefinitions()...).
					Block(
						jen.Id("rets").Op(":=").Id("m").Dot(m.parseReturnsFunctionName()).
							Call(jen.Id("f").Dot("Call").Call(m.generateInputValues()...).Op("...")),
						jen.Return(m.generateConstructorOutputValue()...),
					),
			),
		)
}

func (m *mockedFunctionGenerator) mockerConstructorName() string {
	return "NewMocked" + m.ftype.Name()
}

func (m *mockedFunctionGenerator) parseReturnsFunctionName() string {
	return "parseReturns"
}

func (m *mockedFunctionGenerator) generateInputValues() []jen.Code {
	params := make([]jen.Code, m.ftype.NumIn(), m.ftype.NumIn())
	for i := 0; i < m.ftype.NumIn(); i++ {
		name := fmt.Sprintf("p%d", i+1)
		params[i] = jen.Id(name)
	}
	return params
}

func (m *mockedFunctionGenerator) generateConstructorOutputValue() []jen.Code {
	vals := make([]jen.Code, m.ftype.NumOut(), m.ftype.NumOut())
	for i := 0; i < m.ftype.NumOut(); i++ {
		name := fmt.Sprintf("r%d", i+1)
		vals[i] = jen.Id("rets").Dot(name)
	}
	return vals
}

func (m *mockedFunctionGenerator) generateParseParams() jen.Code {
	body := make([]jen.Code, 0, 0)
	body = append(body, jen.Id("p").Op(":=").Id(m.mockerParamStructName()).Block())
	body = append(body, m.generateParseParamsConversion()...)
	body = append(body, jen.Return(jen.Id("p")))

	return jen.Func().
		Params(jen.Id("m").Op("*").Id(m.mockerStructName())).
		Id(m.parseParamsFunctionName()).
		Params(jen.Id("params").Op("...").Interface()).
		Params(jen.Id(m.mockerParamStructName())).Block(body...)
}

func (m *mockedFunctionGenerator) generateParseParamsConversion() []jen.Code {
	convs := make([]jen.Code, m.ftype.NumIn(), m.ftype.NumIn())

	for i := 0; i < m.ftype.NumIn(); i++ {
		ptype := m.ftype.In(i)
		pname := fmt.Sprintf("p%d", i+1)

		if isTypeNullable(ptype) {
			convs[i] = jen.If(jen.Id("params").Index(jen.Lit(i)).Op("!=").Nil()).Block(
				jen.Id("p").Dot(pname).Op("=").Id("params").Index(jen.Lit(i)).Assert(generateJenFromType(ptype)),
			)
		} else {
			convs[i] = jen.Id("p").Dot(pname).Op("=").Id("params").Index(jen.Lit(i)).Assert(generateJenFromType(ptype))
		}
	}

	return convs
}

func (m *mockedFunctionGenerator) generateParseReturns() jen.Code {
	body := make([]jen.Code, 0, 0)
	body = append(body, jen.Id("r").Op(":=").Id(m.mockerReturnStructName()).Block())
	body = append(body, m.generateParseReturnConversions()...)
	body = append(body, jen.Return(jen.Id("r")))

	return jen.Func().
		Params(jen.Id("m").Op("*").Id(m.mockerStructName())).
		Id("parseReturns").
		Params(jen.Id("returns").Op("...").Interface()).
		Params(jen.Id(m.mockerReturnStructName())).Block(body...)
}

func (m *mockedFunctionGenerator) generateParseReturnConversions() []jen.Code {
	convs := make([]jen.Code, m.ftype.NumOut(), m.ftype.NumOut())
	for i := 0; i < m.ftype.NumOut(); i++ {
		rtype := m.ftype.Out(i)
		pname := fmt.Sprintf("r%d", i+1)
		if isTypeNullable(rtype) {
			convs[i] = jen.If(jen.Id("rets").Index(jen.Lit(i)).Op("!=").Nil()).Block(
				jen.Id("r").Dot(pname).Op("=").Id("rets").Index(jen.Lit(i)).Assert(generateJenFromType(rtype)),
			)
		} else {
			convs[i] = jen.Id("r").Dot(pname).Op("=").Id("rets").Index(jen.Lit(i)).Assert(generateJenFromType(rtype))
		}
	}
	return convs
}

func (m *mockedFunctionGenerator) parseParamsFunctionName() string {
	return "parseParams"
}

func (m *mockedFunctionGenerator) generateMockReturnDefaultValue() jen.Code {
	body := make([]jen.Code, 0, 0)
	for i := 0; i < m.ftype.NumOut(); i++ {
		rname := fmt.Sprintf("r%d", i+1)
		body = append(body, jen.Var().Id(rname).Add(generateJenFromType(m.ftype.Out(i))))
	}

	body = append(
		body,
		jen.Id("m").Dot("mocker").Dot("Mock").Call(
			jen.Id("nTimes"),
			jen.Qual(m.gomockerPackage(), "NewFixedReturnsFuncHandler").Call(m.generateOutputValues()...),
		),
	)

	return jen.Func().
		Params(jen.Id("m").Op("*").Id(m.mockerStructName())).
		Id(m.mockerReturnDefaultValueFuncName()).
		Params(jen.Id("nTimes").Id("int")).Block(body...)
}

func (m *mockedFunctionGenerator) mockerReturnDefaultValueFuncName() string {
	return "MockReturnDefaultValue"
}

func (m *mockedFunctionGenerator) generateOutputValues() []jen.Code {
	returns := make([]jen.Code, m.ftype.NumOut(), m.ftype.NumOut())
	for i := 0; i < m.ftype.NumOut(); i++ {
		name := fmt.Sprintf("r%d", i+1)
		returns[i] = jen.Id(name)
	}
	return returns
}

func (m *mockedFunctionGenerator) generateMockReturnDefaultValueForever() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(m.mockerStructName())).
		Id(m.mockerReturnDefaultValueForeverFuncName()).Params().
		Block(jen.Id("m").Dot(m.mockerReturnDefaultValueFuncName()).Call(
			jen.Qual(m.gomockerPackage(), "LifetimeForever"),
		))
}

func (m *mockedFunctionGenerator) mockerReturnDefaultValueForeverFuncName() string {
	return "MockReturnDefaultValueForever"
}

func (m *mockedFunctionGenerator) generateMockReturnDefaultValueOnce() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(m.mockerStructName())).
		Id(m.mockerReturnDefaultValueOnceFuncName()).Params().
		Block(jen.Id("m").Dot(m.mockerReturnDefaultValueFuncName()).Call(jen.Lit(1)), )
}

func (m *mockedFunctionGenerator) mockerReturnDefaultValueOnceFuncName() string {
	return "MockReturnDefaultValueOnce"
}

func (m *mockedFunctionGenerator) generateMockReturnValues() jen.Code {
	functionParams := make([]jen.Code, 0, 0)
	functionParams = append(functionParams, jen.Id("nTimes").Id("int"))
	functionParams = append(functionParams, m.generateOutputParamDefinitions()...)
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(m.mockerStructName())).
		Id(m.mockReturnValuesFuncName()).
		Params(functionParams...).Block(
		jen.Id("m").Dot("mocker").Dot("Mock").Call(
			jen.Id("nTimes"),
			jen.Qual(m.gomockerPackage(), "NewFixedReturnsFuncHandler").Call(m.generateOutputValues()...),
		),
	)
}

func (m *mockedFunctionGenerator) mockReturnValuesFuncName() string {
	return "MockReturnValues"
}

func (m *mockedFunctionGenerator) generateMockReturnValuesForever() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(m.mockerStructName())).
		Id(m.mockReturnValuesForeverFuncName()).
		Params(m.generateOutputParamDefinitions()...).Block(
		jen.Id("m").Dot(m.mockReturnValuesFuncName()).Call(
			jen.Qual(m.gomockerPackage(), "LifetimeForever"),
			jen.Qual(m.gomockerPackage(), "NewFixedReturnsFuncHandler").Call(m.generateOutputValues()...),
		),
	)
}

func (m *mockedFunctionGenerator) mockReturnValuesForeverFuncName() string {
	return "MockReturnValuesForever"
}

func (m *mockedFunctionGenerator) generateMockReturnValuesOnce() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(m.mockerStructName())).
		Id(m.mockReturnValuesOnceFuncName()).
		Params(m.generateOutputParamDefinitions()...).Block(
		jen.Id("m").Dot(m.mockReturnValuesFuncName()).Call(
			jen.Lit(1),
			jen.Qual(m.gomockerPackage(), "NewFixedReturnsFuncHandler").Call(m.generateOutputValues()...),
		),
	)
}

func (m *mockedFunctionGenerator) mockReturnValuesOnceFuncName() string {
	return "MockReturnValuesOnce"
}

func (m *mockedFunctionGenerator) generateMock() jen.Code {
	params := make([]jen.Code, m.ftype.NumIn(), m.ftype.NumIn())
	for i := 0; i < m.ftype.NumIn(); i++ {
		name := fmt.Sprintf("p%d", i+1)
		params[i] = jen.Id("params").Dot(name)
	}

	return jen.Func().
		Params(jen.Id("m").Op("*").Id(m.mockerStructName())).
		Id(m.mockFuncName()).
		Params(jen.Id("nTimes").Id("int"), jen.Id("f").Add(generateJenFromType(m.ftype))).Block(
		jen.Id("m").Dot("mocker").Dot("Mock").Call(
			jen.Id("nTimes"),
			jen.Qual(m.gomockerPackage(), "NewFuncHandler").Call(
				jen.Func().
					Params(jen.Id("parameters").Op("...").Interface()).
					Params(jen.Index().Interface()).Block(
					jen.Id("params").Op(":=").Id("m").Dot("parseParams").Call(jen.Id("parameters").Op("...")),
					jen.List(m.generateOutputValues()...).Op(":=").Id("f").Call(params...),
					jen.Return(jen.Index().Interface().Values(m.generateOutputValues()...),
					),
				),
			),
		),
	)
}

func (m *mockedFunctionGenerator) mockFuncName() string {
	return "Mock"
}

func (m *mockedFunctionGenerator) generateMockForever() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(m.mockerStructName())).
		Id(m.mockForeverFuncName()).
		Params(jen.Id("f").Add(generateJenFromType(m.ftype))).Block(
		jen.Id("m").Dot(m.mockFuncName()).Call(
			jen.Qual(m.gomockerPackage(), "LifetimeForever"),
			jen.Id("f"),
		),
	)
}

func (m *mockedFunctionGenerator) mockForeverFuncName() string {
	return "MockReturnValuesForever"
}

func (m *mockedFunctionGenerator) generateMockOnce() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(m.mockerStructName())).
		Id(m.mockOnceFuncName()).
		Params(jen.Id("f").Add(generateJenFromType(m.ftype))).Block(
		jen.Id("m").Dot(m.mockFuncName()).Call(jen.Lit(1), jen.Id("f")),
	)
}

func (m *mockedFunctionGenerator) mockOnceFuncName() string {
	return "MockReturnValuesOnce"
}
