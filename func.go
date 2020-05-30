package gomocker

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

type funcMockerGenerator struct {
	namer FuncMockerNamer
}

func (f *funcMockerGenerator) GenerateFunctionMocker(name string, funcType FuncType, withConstructor bool) jen.Code {
	generator := &funcMockerGeneratorHelper{
		funcName:        name,
		funcType:        funcType,
		mockerNamer:     f.namer,
		withConstructor: withConstructor,
	}
	return generator.generate()
}

type funcMockerGeneratorHelper struct {
	funcName        string
	funcType        FuncType
	mockerNamer     FuncMockerNamer
	withConstructor bool
}

func (f *funcMockerGeneratorHelper) generate() jen.Code {
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

func (f *funcMockerGeneratorHelper) generateInvocationStruct() jen.Code {
	return jen.Type().Id(f.getInvocationStructName()).Struct(
		jen.Id("Inputs").Add(f.generateInputStruct()),
		jen.Id("Outputs").Add(f.generateOutputStruct()),
	)
}

func (f *funcMockerGeneratorHelper) getInvocationStructName() string {
	return f.mockerNamer.InvocationName(f.funcName)
}

func (f *funcMockerGeneratorHelper) generateInputStruct() jen.Code {
	return f.generateInputOutputStruct(f.funcType.Inputs, false, f.funcType.IsVariadic)
}

func (f *funcMockerGeneratorHelper) generateInputOutputStruct(
	fields []TypeField,
	isOutput bool,
	isVariadic bool,
) jen.Code {
	paramList := make([]jen.Code, 0, len(fields))
	for i, field := range fields {
		typ := GenerateCode(field.Type, BareFunctionFlag)
		if isVariadic && !isOutput && i == len(fields)-1 {
			typ = jen.Index().Add(typ)
		}
		paramList = append(paramList, jen.Id(makePublic(field.Name)).Add(typ))
	}
	return jen.Struct(paramList...)
}

func (f *funcMockerGeneratorHelper) generateOutputStruct() jen.Code {
	return f.generateInputOutputStruct(f.funcType.Outputs, true, false)
}

func (f *funcMockerGeneratorHelper) generateMockerStruct() jen.Code {
	return jen.Type().Id(f.getMockerStructName()).Struct(
		jen.Id("mux").Qual("sync", "Mutex"),
		jen.Id("handlers").Index().Add(GenerateCode(f.funcType.Type(), BareFunctionFlag)),
		jen.Id("lifetimes").Index().Int(),
		jen.Id("invocations").Index().Id(f.getInvocationStructName()),
	)
}

func (f *funcMockerGeneratorHelper) getMockerStructName() string {
	return f.mockerNamer.MockerName(f.funcName)
}

func (f *funcMockerGeneratorHelper) generateMockMethod() jen.Code {
	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("Mock").
		Params(
			jen.Id("nTimes").Int(),
			jen.Id("f").Add(GenerateCode(f.funcType.Type(), DefaultFlag)),
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

func (f *funcMockerGeneratorHelper) generateLockUnlock() []jen.Code {
	return []jen.Code{
		jen.Id("m").Dot("mux").Dot("Lock").Call(),
		jen.Defer().Id("m").Dot("mux").Dot("Unlock").Call(),
	}
}

func (f *funcMockerGeneratorHelper) alreadyMockedPanicMessage() string {
	return fmt.Sprintf("%s: already mocked forever", f.getMockerStructName())
}

func (f *funcMockerGeneratorHelper) invalidLifetimePanicMessage() string {
	return fmt.Sprintf(
		"%s: invalid lifetime, valid lifetime are positive number and 0 (0 means forever)",
		f.getMockerStructName(),
	)
}

func (f *funcMockerGeneratorHelper) generateMockOnceMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("MockOnce").
		Params(jen.Id("f").Add(GenerateCode(f.funcType.Type(), DefaultFlag))).
		Block(
			jen.Id("m").Dot("Mock").Call(jen.Lit(1), jen.Id("f")),
		)
}

func (f *funcMockerGeneratorHelper) generateMockForeverMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("MockForever").
		Params(jen.Id("f").Add(GenerateCode(f.funcType.Type(), DefaultFlag))).
		Block(
			jen.Id("m").Dot("Mock").Call(jen.Lit(0), jen.Id("f")),
		)
}

func (f *funcMockerGeneratorHelper) generateMockOutputsMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("MockOutputs").
		Params(
			append([]jen.Code{jen.Id("nTimes").Int()}, f.generateOutputParamSignature(true)...)...
		).Block(
		jen.Id("m").Dot("Mock").Call(
			jen.Id("nTimes"),
			jen.Func().
				Params(f.generateInputParamSignature(false)...).
				Params(f.generateOutputParamSignature(false)...).
				Block(jen.Return(f.generateOutputList(true)...)),
		),
	)
}

func (f *funcMockerGeneratorHelper) generateInputParamSignature(withName bool) []jen.Code {
	inputs := make([]jen.Code, 0, len(f.funcType.Inputs))
	for i, field := range f.funcType.Inputs {
		typ := GenerateCode(field.Type, DefaultFlag)
		if f.funcType.IsVariadic && i == len(f.funcType.Inputs)-1 {
			typ = jen.Op("...").Add(typ)
		}
		if withName {
			inputs = append(inputs, jen.Id(field.Name).Add(typ))
		} else {
			inputs = append(inputs, typ)
		}
	}
	return inputs
}

func (f *funcMockerGeneratorHelper) generateOutputParamSignature(withName bool) []jen.Code {
	outputs := make([]jen.Code, 0, len(f.funcType.Outputs))
	for _, field := range f.funcType.Outputs {
		if withName {
			outputs = append(outputs, jen.Id(field.Name).Add(GenerateCode(field.Type, DefaultFlag)))
		} else {
			outputs = append(outputs, GenerateCode(field.Type, DefaultFlag))
		}
	}
	return outputs
}

func (f *funcMockerGeneratorHelper) generateOutputList(useFieldName bool) []jen.Code {
	outputs := make([]jen.Code, 0, len(f.funcType.Outputs))
	for i, field := range f.funcType.Outputs {
		if useFieldName {
			outputs = append(outputs, jen.Id(field.Name))
		} else {
			outputs = append(outputs, jen.Id(fmt.Sprint("out", i+1)))
		}
	}
	return outputs
}

func (f *funcMockerGeneratorHelper) generateMockOutputsOnceMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("MockOutputsOnce").
		Params(f.generateOutputParamSignature(true)...).
		Block(
			jen.Id("m").Dot("MockOutputs").Call(
				append([]jen.Code{jen.Lit(1)}, f.generateOutputList(true)...)...
			),
		)
}

func (f *funcMockerGeneratorHelper) generateMockOutputsForeverMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("MockOutputsForever").
		Params(f.generateOutputParamSignature(true)...).
		Block(
			jen.Id("m").Dot("MockOutputs").Call(
				append([]jen.Code{jen.Lit(0)}, f.generateOutputList(true)...)...
			),
		)
}

func (f *funcMockerGeneratorHelper) generateMockDefaultsMethod() jen.Code {
	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("MockDefaults").
		Params(jen.Id("nTimes").Int())

	body := make([]jen.Code, 0, 0)

	for i, field := range f.funcType.Outputs {
		body = append(
			body,
			jen.Var().Add(jen.Id(fmt.Sprint("out", i+1))).Add(GenerateCode(field.Type, BareFunctionFlag)),
		)
	}

	body = append(body, jen.Id("m").Dot("MockOutputs").Call(append(
		[]jen.Code{jen.Id("nTimes")},
		f.generateOutputList(false)...,
	)...))

	return code.Block(body...)
}

func (f *funcMockerGeneratorHelper) generateMockDefaultsOnce() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("MockDefaultsOnce").
		Params().
		Block(jen.Id("m").Dot("MockDefaults").Call(jen.Lit(1)))
}

func (f *funcMockerGeneratorHelper) generateMockDefaultsForever() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("MockDefaultsForever").
		Params().
		Block(jen.Id("m").Dot("MockDefaults").Call(jen.Lit(0)))
}

func (f *funcMockerGeneratorHelper) generateCallMethod() jen.Code {
	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("Call").
		Params(f.generateInputParamSignature(true)...).
		Params(f.generateOutputParamSignature(false)...)

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

	if hasOutput := len(f.funcType.Outputs) > 0; hasOutput {
		body = append(
			body,
			jen.List(f.generateOutputList(false)...).Op(":=").Id("handler").Call(f.generateInputList(true)...),
		)
	} else {
		body = append(body, jen.Id("handler").Call(f.generateInputList(true)...))
	}

	body = append(
		body,
		jen.Id("input").Op(":=").Add(f.generateInputStruct()).Values(f.generateInputList(false)...),
		jen.Id("output").Op(":=").Add(f.generateOutputStruct()).Values(f.generateOutputList(false)...),
		jen.Id("invoc").Op(":=").Id(f.getInvocationStructName()).Values(jen.Id("input"), jen.Id("output")),
		jen.Id("m").Dot("invocations").Op("=").Append(jen.Id("m").Dot("invocations"), jen.Id("invoc")),
	)

	body = append(body, jen.Return(f.generateOutputList(false)...))

	code.Block(body...)
	return code
}

func (f *funcMockerGeneratorHelper) noHandler() string {
	return fmt.Sprintf("%s: no handler", f.getMockerStructName())
}

func (f *funcMockerGeneratorHelper) generateInputList(considerEllipsis bool) []jen.Code {
	inputs := make([]jen.Code, 0, len(f.funcType.Inputs))
	for i, field := range f.funcType.Inputs {
		if considerEllipsis && f.funcType.IsVariadic && i == len(f.funcType.Inputs)-1 {
			inputs = append(inputs, jen.Id(field.Name).Op("..."))
		} else {
			inputs = append(inputs, jen.Id(field.Name))
		}
	}
	return inputs
}

func (f *funcMockerGeneratorHelper) generateInvocationsMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("Invocations").
		Params().
		Params(jen.Index().Id(f.getInvocationStructName())).
		Block(jen.Return(jen.Id("m").Dot("invocations")))
}

func (f *funcMockerGeneratorHelper) generateTakeOneInvocationMethod() jen.Code {
	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("TakeOneInvocation").
		Params().
		Params(jen.Id(f.getInvocationStructName()))

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

	return code.Block(body...)
}

func (f *funcMockerGeneratorHelper) noInvocationPanicMessage() string {
	return fmt.Sprintf("%s: no invocations", f.getMockerStructName())
}

func (f *funcMockerGeneratorHelper) generateFuncMockerConstructor() jen.Code {
	return jen.Func().
		Id(f.getMockerConstructorName()).
		Params().
		Params(
			GenerateCode(f.funcType.Type(), BareFunctionFlag),
			jen.Op("*").Id(f.getMockerStructName()),
		).
		Block(
			jen.Id("m").Op(":=").Op("&").Id(f.getMockerStructName()).Values(),
			jen.Return(jen.Id("m").Dot("Call"), jen.Id("m")),
		)
}

func (f *funcMockerGeneratorHelper) getMockerConstructorName() string {
	return f.mockerNamer.ConstructorName(f.funcName)
}
