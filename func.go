package gomocker

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/dave/jennifer/jen"
)

// TODO (jauhararifin): make all local variable have prefix
// TODO (jauhararifin): check if Namer returns wrong name (when collide with local var, when not capitalized)

func GenerateFuncMocker(t reflect.Type, name string) (jen.Code, error) {
	if t == nil {
		return nil, errors.New("invalid type")
	}
	if t.Kind() != reflect.Func {
		return nil, errors.New("type must a function")
	}

	generator := &funcMockerGenerator{
		funcType:           t,
		name:               name,
		mockerName:         name + "Mocker",
		inputNamer:         &defaultInputNamer{},
		outputNamer:        &defaultOutputNamer{},
		includeConstructor: true,
	}
	return generator.generate(), nil
}

type Namer interface {
	Name(t reflect.Type, i int, public bool) string
}

type defaultInputNamer struct{}

func (*defaultInputNamer) Name(t reflect.Type, i int, public bool) string {
	if public {
		return fmt.Sprintf("Arg%d", i+1)
	}
	return fmt.Sprintf("arg%d", i+1)
}

type defaultOutputNamer struct{}

func (*defaultOutputNamer) Name(t reflect.Type, i int, public bool) string {
	if public {
		return fmt.Sprintf("Out%d", i+1)
	}
	return fmt.Sprintf("out%d", i+1)
}

type funcMockerGenerator struct {
	funcType                reflect.Type
	name                    string
	mockerName              string
	inputNamer, outputNamer Namer
	includeConstructor      bool
}

func (f *funcMockerGenerator) generate() jen.Code {
	code := jen.
		Add(f.generateMockerStruct()).Line().Line().
		Add(f.generateMockMethod()).Line().Line().
		Add(f.generateMockOnceMethod()).Line().Line().
		Add(f.generateMockForeverMethod()).Line().Line().
		Add(f.generateMockOutputsMethod()).Line().Line().
		Add(f.generateMockOutputsOnceMethod()).Line().Line().
		Add(f.generateMockOutputsForeverMethod()).Line().Line().
		Add(f.generateMockDefaultsMethod()).Line().Line().
		Add(f.generateMockDefaultsOnce()).Line().Line().
		Add(f.generateMockDefaultsForever()).Line().Line().
		Add(f.generateCallMethod()).Line().Line().
		Add(f.generateInvocationsMethod()).Line().Line().
		Add(f.generateTakeOneInvocationMethod()).Line().Line()
	if f.includeConstructor {
		code.Add(f.generateFuncMockerConstructor()).Line().Line()
	}
	return code
}

func (f *funcMockerGenerator) generateMockerStruct() jen.Code {
	return jen.Type().Id(f.mockerName).Struct(
		jen.Id("mux").Qual("sync", "Mutex"),
		jen.Id("handlers").Index().Add(generateDefinitionFromType(f.funcType)),
		jen.Id("lifetimes").Index().Int(),
		jen.Id("invocations").Index().Add(f.generateInvocationStructType()),
	)
}

func (f *funcMockerGenerator) generateInvocationStructType() jen.Code {
	return jen.Struct(
		jen.Id("Inputs").Add(f.generateInputStruct()),
		jen.Id("Outputs").Add(f.generateOutputStruct()),
	)
}

func (f *funcMockerGenerator) generateInputStruct() jen.Code {
	nInput := f.funcType.NumIn()
	inputFields := make([]jen.Code, nInput, nInput)
	for i := 0; i < nInput; i++ {
		inputFields = append(
			inputFields,
			jen.Id(f.inputNamer.Name(f.funcType, i, true)).Add(generateDefinitionFromType(f.funcType.In(i))),
		)
	}
	return jen.Struct(inputFields...)
}

func (f *funcMockerGenerator) generateOutputStruct() jen.Code {
	nOutput := f.funcType.NumOut()
	outputFields := make([]jen.Code, nOutput, nOutput)
	for i := 0; i < nOutput; i++ {
		outputFields = append(
			outputFields,
			jen.Id(f.outputNamer.Name(f.funcType, i, true)).Add(generateDefinitionFromType(f.funcType.Out(i))),
		)
	}
	return jen.Struct(outputFields...)
}

func (f *funcMockerGenerator) generateMockMethod() jen.Code {
	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.mockerName)).
		Id("Mock").
		Params(
			jen.Id("nTimes").Int(),
			jen.Id("f").Add(generateDefinitionFromType(f.funcType)),
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

	code.Block(body...)
	return code
}

func (f *funcMockerGenerator) generateLockUnlock() []jen.Code {
	return []jen.Code{
		jen.Id("m").Dot("mux").Dot("Lock").Call(),
		jen.Defer().Id("m").Dot("mux").Dot("Unlock").Call(),
	}
}

func (f *funcMockerGenerator) alreadyMockedPanicMessage() string {
	return fmt.Sprintf("%s: already mocked forever", f.mockerName)
}

func (f *funcMockerGenerator) invalidLifetimePanicMessage() string {
	return fmt.Sprintf("%s: invalid lifetime, valid lifetime are positive number and 0 (0 means forever)", f.mockerName)
}

func (f *funcMockerGenerator) generateMockOnceMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.mockerName)).
		Id("MockOnce").
		Params(jen.Id("f").Add(generateDefinitionFromType(f.funcType))).
		Block(
			jen.Id("m").Dot("Mock").Call(jen.Lit(1), jen.Id("f")),
		)
}

func (f *funcMockerGenerator) generateMockForeverMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.mockerName)).
		Id("MockForever").
		Params(jen.Id("f").Add(generateDefinitionFromType(f.funcType))).
		Block(
			jen.Id("m").Dot("Mock").Call(jen.Lit(0), jen.Id("f")),
		)
}

func (f *funcMockerGenerator) generateMockOutputsMethod() jen.Code {
	inputs, outputs := f.generateParamDef(true)
	_, _, outputList := f.generateParamList()
	_, innerOutput := f.generateParamDef(false)
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.mockerName)).
		Id("MockOutputs").
		Params(
			append([]jen.Code{jen.Id("nTimes").Int()}, outputs...)...
		).Block(
		jen.Id("m").Dot("Mock").Call(
			jen.Id("nTimes"),
			jen.Func().Params(inputs...).Params(innerOutput...).Block(jen.Return(outputList...)),
		),
	)
}

func (f *funcMockerGenerator) generateMockOutputsOnceMethod() jen.Code {
	_, outputs := f.generateParamDef(true)
	_, _, outputList := f.generateParamList()
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.mockerName)).
		Id("MockOutputsOnce").
		Params(outputs...).
		Block(
			jen.Id("m").Dot("MockOutputs").Call(
				append([]jen.Code{jen.Lit(1)}, outputList...)...
			),
		)
}

func (f *funcMockerGenerator) generateMockOutputsForeverMethod() jen.Code {
	_, outputs := f.generateParamDef(true)
	_, _, outputList := f.generateParamList()
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.mockerName)).
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
		Params(jen.Id("m").Op("*").Id(f.mockerName)).
		Id("MockDefaults").
		Params(jen.Id("nTimes").Int())

	body := make([]jen.Code, 0, 0)

	_, _, outputList := f.generateParamList()
	for i, o := range outputList {
		body = append(
			body,
			jen.Var().Add(o).Add(generateDefinitionFromType(f.funcType.Out(i))),
		)
	}

	body = append(
		body,
		jen.Id("m").Dot("MockOutputs").Call(
			append(
				[]jen.Code{jen.Id("nTimes")},
				outputList...,
			)...,
		),
	)

	code.Block(body...)
	return code
}

func (f *funcMockerGenerator) generateMockDefaultsOnce() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.mockerName)).
		Id("MockDefaultsOnce").
		Params().
		Block(jen.Id("m").Dot("MockDefaults").Call(jen.Lit(1)))
}

func (f *funcMockerGenerator) generateMockDefaultsForever() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.mockerName)).
		Id("MockDefaultsForever").
		Params().
		Block(jen.Id("m").Dot("MockDefaults").Call(jen.Lit(0)))
}

func (f *funcMockerGenerator) generateCallMethod() jen.Code {
	params, returns := f.generateParamDef(true)
	code := jen.Func().Params(jen.Id("m").Op("*").Id(f.mockerName)).Id("Call").Params(params...).Params(returns...)

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

	body = append(
		body,
		jen.Id("input").Op(":=").Add(f.generateInputStruct()).Values(inputs...),
		jen.Id("output").Op(":=").Add(f.generateOutputStruct()).Values(outputs...),
		jen.Id("invoc").Op(":=").Add(f.generateInvocationStructType()).Values(jen.Id("input"), jen.Id("output")),
		jen.Id("m").Dot("invocations").Op("=").Append(jen.Id("m").Dot("invocations"), jen.Id("invoc")),
	)

	body = append(body, jen.Return(outputs...))

	code.Block(body...)
	return code
}

func (f *funcMockerGenerator) noHandler() string {
	return fmt.Sprintf("%s: no handler", f.mockerName)
}

func (f *funcMockerGenerator) generateParamDef(withOutputName bool) (inputs, outputs []jen.Code) {
	nIn := f.funcType.NumIn()
	inputs = make([]jen.Code, 0, nIn)
	for i := 0; i < nIn; i++ {
		c := jen.Id(f.inputNamer.Name(f.funcType, i, false))
		in := f.funcType.In(i)
		if i == nIn-1 && f.funcType.IsVariadic() {
			inputs = append(inputs, c.Op("...").Add(generateDefinitionFromType(in.Elem())))
		} else {
			inputs = append(inputs, c.Add(generateDefinitionFromType(in)))
		}
	}

	nOut := f.funcType.NumOut()
	outputs = make([]jen.Code, 0, nOut)
	for i := 0; i < nOut; i++ {
		out := f.funcType.Out(i)
		c := jen.Empty()
		if withOutputName {
			c = jen.Id(f.outputNamer.Name(f.funcType, i, false))
		}
		outputs = append(outputs, c.Add(generateDefinitionFromType(out)))
	}
	return inputs, outputs
}

func (f *funcMockerGenerator) generateParamList() (inputs, inputsForCall, outputs []jen.Code) {
	nIn := f.funcType.NumIn()
	inputsForCall = make([]jen.Code, 0, nIn)
	inputs = make([]jen.Code, 0, nIn)
	for i := 0; i < nIn; i++ {
		callParam := jen.Id(f.inputNamer.Name(f.funcType, i, false))
		if i == nIn-1 && f.funcType.IsVariadic() {
			callParam.Op("...")
		}
		inputsForCall = append(inputsForCall, callParam)
		inputs = append(inputs, jen.Id(f.inputNamer.Name(f.funcType, i, false)))
	}

	nOut := f.funcType.NumOut()
	outputs = make([]jen.Code, 0, nOut)
	for i := 0; i < nOut; i++ {
		outputs = append(outputs, jen.Id(f.outputNamer.Name(f.funcType, i, false)))
	}

	return
}

func (f *funcMockerGenerator) generateInvocationsMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.mockerName)).
		Id("Invocations").
		Params().
		Params(jen.Index().Add(f.generateInvocationStructType())).
		Block(jen.Return(jen.Id("m").Dot("invocations")))
}

func (f *funcMockerGenerator) generateTakeOneInvocationMethod() jen.Code {
	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.mockerName)).
		Id("TakeOneInvocation").
		Params().
		Params(f.generateInvocationStructType())

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
	return fmt.Sprintf("%s: no invocations", f.mockerName)
}

func (f *funcMockerGenerator) generateFuncMockerConstructor() jen.Code {
	return jen.Func().
		Id(fmt.Sprintf("MakeMocked%s", f.name)).
		Params().
		Params(
			generateDefinitionFromType(f.funcType),
			jen.Op("*").Id(f.mockerName),
		).
		Block(
			jen.Id("m").Op(":=").Op("&").Id(f.mockerName).Values(),
			jen.Return(jen.Id("m").Dot("Call"), jen.Id("m")),
		)
}
