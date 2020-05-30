package gomocker

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/dave/jennifer/jen"
)

type funcMockerGenerator struct {
	funcName        string
	funcType        FuncType
	mockerNamer     FuncMockerNamer
	withConstructor bool
}

func (f *funcMockerGenerator) generate() jen.Code {
	steps := []stepFunc{
		f.generateInvocationStruct,
		f.generateMockerStruct,
		f.generateMockMethod,
		f.generateMockOnceMethod,
		f.generateMockForeverMethod,
		f.generateMockOutputsMethod,
		f.generateMockOutputsOnceMethod,
		f.generateMockOutputsForeverMethod,
		f.generateMockDefaultsMethod,
		f.generateMockDefaultsOnce,
		f.generateMockDefaultsForever,
		f.generateCallMethod,
		f.generateInvocationsMethod,
		f.generateTakeOneInvocationMethod,
	}
	if f.withConstructor {
		steps = append(steps, f.generateFuncMockerConstructor)
	}
	return concatSteps(steps...)
}

func (f *funcMockerGenerator) generateMockerStruct() jen.Code {
	return jen.Type().Id(f.getMockerName()).Struct(
		jen.Id("mux").Qual("sync", "Mutex"),
		jen.Id("handlers").Index().Add(GenerateCode(f.funcType.Type(), BareFunctionFlag)),
		jen.Id("lifetimes").Index().Int(),
		jen.Id("invocations").Index().Id(f.getInvocationStructName()),
	)
}

func (f *funcMockerGenerator) getMockerName() string {
	return f.mockerNamer.MockerName(f.funcName)
}

func (f *funcMockerGenerator) generateInvocationStruct() jen.Code {
	return jen.Type().Id(f.getInvocationStructName()).Struct(
		jen.Id("Inputs").Add(f.generateInputStruct()),
		jen.Id("Outputs").Add(f.generateOutputStruct()),
	)
}

func (f *funcMockerGenerator) getInvocationStructName() string {
	return f.mockerNamer.InvocationName(f.funcName)
}

func (f *funcMockerGenerator) generateInputStruct() jen.Code {
	return f.generateInputOutputStruct(f.funcType.Inputs, false, f.funcType.IsVariadic)
}

func (f *funcMockerGenerator) generateInputOutputStruct(fields []TypeField, isOutput bool, isVariadic bool) jen.Code {
	paramList := make([]jen.Code, 0, len(fields))
	for i, field := range fields {
		typ := GenerateCode(field.Type, BareFunctionFlag)
		if isVariadic && !isOutput && i == len(fields)-1 {
			typ = jen.Index().Add(typ)
		}
		paramList = append(paramList, jen.Id(strings.Title(field.Name)).Add(typ))
	}
	return jen.Struct(paramList...)
}

type paramName struct {
	name string
	expr ast.Expr
}

func (f *funcMockerGenerator) generateOutputStruct() jen.Code {
	return f.generateInputOutputStruct(f.funcType.Outputs, true, false)
}

func (f *funcMockerGenerator) generateMockMethod() jen.Code {
	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("Mock").
		Params(
			jen.Id("nTimes").Int(),
			jen.Id("f").Add(GenerateCode(f.funcType.Type(), 0)),
		)

	body := f.generateLockUnlock()

	body = append(
		body,
		jen.Id("nHandler").Op(":=").Len(jen.Id("m").Dot("lifetimes")),
		jen.If(
			jen.Id("nHandler").Op(">").Lit(0).
				Op("&&").
				Id("m").Dot("lifetimes").Index(jen.Id("nHandler").Op("-").Lit(1)).Op("==").Lit(0),
		).Block(
			jen.Panic(jen.Lit(f.alreadyMockedPanicMessage())),
		),
	)

	body = append(
		body,
		jen.If(jen.Id("nTimes").Op("<").Lit(0)).Block(
			jen.Panic(jen.Lit(f.invalidLifetimePanicMessage())),
		),
	)

	body = append(
		body,
		jen.Id("m").Dot("handlers").Op("=").Append(jen.Id("m").Dot("handlers"), jen.Id("f")),
		jen.Id("m").Dot("lifetimes").Op("=").Append(jen.Id("m").Dot("lifetimes"), jen.Id("nTimes")),
	)

	return code.Block(body...)
}

func (f *funcMockerGenerator) generateLockUnlock() []jen.Code {
	return []jen.Code{
		jen.Id("m").Dot("mux").Dot("Lock").Call(),
		jen.Defer().Id("m").Dot("mux").Dot("Unlock").Call(),
	}
}

func (f *funcMockerGenerator) alreadyMockedPanicMessage() string {
	return fmt.Sprintf("%s: already mocked forever", f.getMockerName())
}

func (f *funcMockerGenerator) invalidLifetimePanicMessage() string {
	return fmt.Sprintf("%s: invalid lifetime, valid lifetime are positive number and 0 (0 means forever)", f.getMockerName())
}

func (f *funcMockerGenerator) generateMockOnceMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockOnce").
		Params(jen.Id("f").Add(GenerateCode(f.funcType.Type(), 0))).
		Block(
			jen.Id("m").Dot("Mock").Call(jen.Lit(1), jen.Id("f")),
		)
}

func (f *funcMockerGenerator) generateMockForeverMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockForever").
		Params(jen.Id("f").Add(GenerateCode(f.funcType.Type(), 0))).
		Block(
			jen.Id("m").Dot("Mock").Call(jen.Lit(0), jen.Id("f")),
		)
}

func (f *funcMockerGenerator) generateMockOutputsMethod() jen.Code {
	_, outputs := f.generateParamSignature(true, true)
	_, _, outputList := f.generateParamList()
	innerInput, innerOutput := f.generateParamSignature(false, false)

	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockOutputs").
		Params(
			append([]jen.Code{jen.Id("nTimes").Int()}, outputs...)...
		).Block(
		jen.Id("m").Dot("Mock").Call(
			jen.Id("nTimes"),
			jen.Func().Params(innerInput...).Params(innerOutput...).Block(jen.Return(outputList...)),
		),
	)
}

func (f *funcMockerGenerator) generateParamSignature(withOutputName bool, withParamName bool) (inputs, outputs []jen.Code) {
	inputs = make([]jen.Code, 0, len(f.funcType.Inputs))
	for i, field := range f.funcType.Inputs {
		typ := GenerateCode(field.Type, 0)
		if f.funcType.IsVariadic && i == len(f.funcType.Inputs)-1 {
			typ = jen.Op("...").Add(typ)
		}
		if withParamName {
			inputs = append(inputs, jen.Id(field.Name).Add(typ))
		} else {
			inputs = append(inputs, typ)
		}
	}

	outputs = make([]jen.Code, 0, len(f.funcType.Outputs))
	for _, field := range f.funcType.Outputs {
		if withParamName {
			outputs = append(outputs, jen.Id(field.Name).Add(GenerateCode(field.Type, 0)))
		} else {
			outputs = append(outputs, GenerateCode(field.Type, 0))
		}
	}

	return inputs, outputs
}

func (f *funcMockerGenerator) generateParamList() (inputs, inputsForCall, outputs []jen.Code) {
	inputsForCall = make([]jen.Code, 0, len(f.funcType.Inputs))
	inputs = make([]jen.Code, 0, len(f.funcType.Inputs))
	for i, field := range f.funcType.Inputs {
		callParam := jen.Id(field.Name)
		if f.funcType.IsVariadic && i == len(f.funcType.Inputs)-1 {
			callParam.Op("...")
		}
		inputsForCall = append(inputsForCall, callParam)
		inputs = append(inputs, jen.Id(field.Name))
	}

	outputs = make([]jen.Code, 0, len(f.funcType.Outputs))
	for _, field := range f.funcType.Outputs {
		outputs = append(outputs, jen.Id(field.Name))
	}

	return
}

func (f *funcMockerGenerator) generateMockOutputsOnceMethod() jen.Code {
	_, outputs := f.generateParamSignature(true, true)
	_, _, outputList := f.generateParamList()
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockOutputsOnce").
		Params(outputs...).
		Block(
			jen.Id("m").Dot("MockOutputs").Call(
				append([]jen.Code{jen.Lit(1)}, outputList...)...
			),
		)
}

func (f *funcMockerGenerator) generateMockOutputsForeverMethod() jen.Code {
	_, outputs := f.generateParamSignature(true, true)
	_, _, outputList := f.generateParamList()
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockOutputsForever").
		Params(outputs...).
		Block(
			jen.Id("m").Dot("MockOutputs").Call(
				append([]jen.Code{jen.Lit(0)}, outputList...)...
			),
		)
}

func (f *funcMockerGenerator) generateMockDefaultsMethod() jen.Code {
	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockDefaults").
		Params(jen.Id("nTimes").Int())

	body := make([]jen.Code, 0, 0)

	for _, field := range f.funcType.Outputs {
		body = append(
			body,
			jen.Var().Add(jen.Id(field.Name)).Add(GenerateCode(field.Type, BareFunctionFlag)),
		)
	}

	mockOutputParams := []jen.Code{jen.Id("nTimes")}
	for _, field := range f.funcType.Outputs {
		mockOutputParams = append(mockOutputParams, jen.Id(field.Name))
	}
	body = append(body, jen.Id("m").Dot("MockOutputs").Call(mockOutputParams...))
	code.Block(body...)

	return code
}

func (f *funcMockerGenerator) generateMockDefaultsOnce() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockDefaultsOnce").
		Params().
		Block(jen.Id("m").Dot("MockDefaults").Call(jen.Lit(1)))
}

func (f *funcMockerGenerator) generateMockDefaultsForever() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockDefaultsForever").
		Params().
		Block(jen.Id("m").Dot("MockDefaults").Call(jen.Lit(0)))
}

func (f *funcMockerGenerator) generateCallMethod() jen.Code {
	params, returns := f.generateParamSignature(true, true)

	code := jen.Func().Params(jen.Id("m").Op("*").Id(f.getMockerName())).Id("Call").Params(params...).Params(returns...)

	body := f.generateLockUnlock()

	body = append(
		body,
		jen.If(jen.Len(jen.Id("m").Dot("handlers")).Op("==").Lit(0)).Block(jen.Panic(jen.Lit(f.noHandler()))),
	)

	body = append(
		body,
		jen.Id("handler").Op(":=").Id("m").Dot("handlers").Index(jen.Lit(0)),
		jen.If(jen.Id("m").Dot("lifetimes").Index(jen.Lit(0))).Op("==").Lit(1).Block(
			jen.Id("m").Dot("handlers").Op("=").Id("m").Dot("handlers").Index(jen.Lit(1), jen.Empty()),
			jen.Id("m").Dot("lifetimes").Op("=").Id("m").Dot("lifetimes").Index(jen.Lit(1), jen.Empty()),
		).Else().If(jen.Id("m").Dot("lifetimes").Index(jen.Lit(0)).Op(">").Lit(1)).Block(
			jen.Id("m").Dot("lifetimes").Index(jen.Lit(0)).Op("--"),
		),
	)

	inputs, inputsForCall, outputs := f.generateParamList()

	if len(outputs) > 0 {
		body = append(body, jen.List(outputs...).Op("=").Id("handler").Call(inputsForCall...))
	} else {
		body = append(body, jen.Id("handler").Call(inputsForCall...))
	}

	inputStruct := f.generateInputStruct()
	outputStruct := f.generateOutputStruct()
	invocStruct := f.getInvocationStructName()

	body = append(
		body,
		jen.Id("input").Op(":=").Add(inputStruct).Values(inputs...),
		jen.Id("output").Op(":=").Add(outputStruct).Values(outputs...),
		jen.Id("invoc").Op(":=").Id(invocStruct).Values(jen.Id("input"), jen.Id("output")),
		jen.Id("m").Dot("invocations").Op("=").Append(jen.Id("m").Dot("invocations"), jen.Id("invoc")),
	)

	body = append(body, jen.Return(outputs...))

	code.Block(body...)
	return code
}

func (f *funcMockerGenerator) noHandler() string {
	return fmt.Sprintf("%s: no handler", f.getMockerName())
}

func (f *funcMockerGenerator) generateInvocationsMethod() jen.Code {
	invocStruct := f.getInvocationStructName()
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("Invocations").
		Params().
		Params(jen.Index().Id(invocStruct)).
		Block(jen.Return(jen.Id("m").Dot("invocations")))
}

func (f *funcMockerGenerator) generateTakeOneInvocationMethod() jen.Code {
	invocStruct := f.getInvocationStructName()
	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("TakeOneInvocation").
		Params().
		Params(jen.Id(invocStruct))

	body := f.generateLockUnlock()

	body = append(
		body,
		jen.If(jen.Len(jen.Id("m").Dot("invocations")).Op("==").Lit(0)).
			Block(jen.Panic(jen.Lit(f.noInvocationPanicMessage()))),
	)

	body = append(
		body,
		jen.Id("invoc").Op(":=").Id("m").Dot("invocations").Index(jen.Lit(0)),
		jen.Id("m").Dot("invocations").Op("=").Id("m").Dot("invocations").Index(jen.Lit(1), jen.Empty()),
		jen.Return(jen.Id("invoc")),
	)

	code.Block(body...)
	return code
}

func (f *funcMockerGenerator) noInvocationPanicMessage() string {
	return fmt.Sprintf("%s: no invocations", f.getMockerName())
}

func (f *funcMockerGenerator) generateFuncMockerConstructor() jen.Code {
	return jen.Func().
		Id(f.getMockerConstructorName()).
		Params().
		Params(
			GenerateCode(f.funcType.Type(), BareFunctionFlag),
			jen.Op("*").Id(f.getMockerName()),
		).
		Block(
			jen.Id("m").Op(":=").Op("&").Id(f.getMockerName()).Values(),
			jen.Return(jen.Id("m").Dot("Call"), jen.Id("m")),
		)
}

func (f *funcMockerGenerator) getMockerConstructorName() string {
	return f.mockerNamer.ConstructorName(f.funcName)
}
