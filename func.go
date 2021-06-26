package gomocker

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/jauhararifin/gotype"
)

type funcMockerGenerator struct {
	namer FuncMockerNamer
}

func (f *funcMockerGenerator) GenerateFunctionMocker(name string, funcType gotype.FuncType, withConstructor bool) jen.Code {
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
	funcType        gotype.FuncType
	mockerNamer     FuncMockerNamer
	withConstructor bool
}

func (f *funcMockerGeneratorHelper) generate() jen.Code {
	return jen.CustomFunc(jen.Options{}, func(g *jen.Group) {
		g.Add(f.generateInvocationStruct()).Line().Line()
		g.Add(f.generateMockerStruct()).Line().Line()
		g.Add(f.generateMockMethod()).Line().Line()
		g.Add(f.generateMockOnceMethod()).Line().Line()
		g.Add(f.generateMockForeverMethod()).Line().Line()
		g.Add(f.generateMockOutputsMethod()).Line().Line()
		g.Add(f.generateMockOutputsOnceMethod()).Line().Line()
		g.Add(f.generateMockOutputsForeverMethod()).Line().Line()
		g.Add(f.generateMockDefaultsMethod()).Line().Line()
		g.Add(f.generateMockDefaultsOnce()).Line().Line()
		g.Add(f.generateMockDefaultsForever()).Line().Line()
		g.Add(f.generateCallMethod()).Line().Line()
		g.Add(f.generateInvocationsMethod()).Line().Line()
		g.Add(f.generateTakeOneInvocationMethod()).Line().Line()

		if f.withConstructor {
			g.Add(f.generateFuncMockerConstructor()).Line().Line()
		}
	})
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
	fields []gotype.TypeField,
	isOutput bool,
	isVariadic bool,
) jen.Code {
	return jen.StructFunc(func(g *jen.Group) {
		for i, field := range fields {
			typ := GenerateCode(field.Type, BareFunctionFlag)
			if isVariadic && !isOutput && i == len(fields)-1 {
				typ = jen.Index().Add(typ)
			}
			g.Id(makePublic(field.Name)).Add(typ)
		}
	})
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
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("Mock").
		Params(
			jen.Id("nTimes").Int(),
			jen.Id("f").Add(GenerateCode(f.funcType.Type(), DefaultFlag)),
		).
		BlockFunc(func(g *jen.Group) {
			f.generateLockUnlock(g)

			g.Id("nHandler").Op(":=").Len(jen.Id("m").Dot("lifetimes"))
			g.If(
				jen.Id("nHandler").Op(">").Lit(0).
					Op("&&").
					Id("m").Dot("lifetimes").Index(jen.Id("nHandler").Op("-").Lit(1)).Op("==").Lit(0),
			).Block(
				jen.Panic(jen.Lit(f.alreadyMockedPanicMessage())),
			)

			g.If(jen.Id("nTimes").Op("<").Lit(0)).Block(
				jen.Panic(jen.Lit(f.invalidLifetimePanicMessage())),
			)

			g.Id("m").Dot("handlers").Op("=").Append(jen.Id("m").Dot("handlers"), jen.Id("f"))
			g.Id("m").Dot("lifetimes").Op("=").Append(jen.Id("m").Dot("lifetimes"), jen.Id("nTimes"))
		})
}

func (f *funcMockerGeneratorHelper) generateLockUnlock(g *jen.Group) {
	g.Id("m").Dot("mux").Dot("Lock").Call()
	g.Defer().Id("m").Dot("mux").Dot("Unlock").Call()
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
			append([]jen.Code{jen.Id("nTimes").Int()}, f.generateOutputParamSignature(true)...)...,
		).Block(
		jen.Id("m").Dot("Mock").Call(
			jen.Id("nTimes"),
			jen.Func().
				Params(f.generateInputParamSignature(false)...).
				Params(f.generateOutputParamSignature(false)...).
				Block(jen.ReturnFunc(f.generateOutputListWithName)),
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

func (f *funcMockerGeneratorHelper) generateOutputListWithName(g *jen.Group) {
	for _, field := range f.funcType.Outputs {
		g.Id(field.Name)
	}
}

func (f *funcMockerGeneratorHelper) generateOutputListWithoutName(g *jen.Group) {
	for i := range f.funcType.Outputs {
		g.Id(fmt.Sprint("out", i+1))
	}
}

func (f *funcMockerGeneratorHelper) generateMockOutputsOnceMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("MockOutputsOnce").
		Params(f.generateOutputParamSignature(true)...).
		Block(
			jen.Id("m").Dot("MockOutputs").CallFunc(func(g *jen.Group) {
				g.Lit(1)
				f.generateOutputListWithName(g)
			}),
		)
}

func (f *funcMockerGeneratorHelper) generateMockOutputsForeverMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("MockOutputsForever").
		Params(f.generateOutputParamSignature(true)...).
		Block(
			jen.Id("m").Dot("MockOutputs").CallFunc(func(g *jen.Group) {
				g.Lit(0)
				f.generateOutputListWithName(g)
			}),
		)
}

func (f *funcMockerGeneratorHelper) generateMockDefaultsMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("MockDefaults").
		Params(jen.Id("nTimes").Int()).
		BlockFunc(func(g *jen.Group) {
			for i, field := range f.funcType.Outputs {
				g.Var().Add(jen.Id(fmt.Sprint("out", i+1))).Add(GenerateCode(field.Type, BareFunctionFlag))
			}

			g.Id("m").Dot("MockOutputs").CallFunc(func(g *jen.Group) {
				g.Id("nTimes")
				f.generateOutputListWithoutName(g)
			})
		})
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
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("Call").
		Params(f.generateInputParamSignature(true)...).
		Params(f.generateOutputParamSignature(false)...).
		BlockFunc(func(g *jen.Group) {
			f.generateLockUnlock(g)

			g.If(jen.Len(jen.Id("m").Dot("handlers")).Op("==").Lit(0)).Block(jen.Panic(jen.Lit(f.noHandler())))

			g.Id("handler").Op(":=").Id("m").Dot("handlers").Index(jen.Lit(0))
			g.If(jen.Id("m").Dot("lifetimes").Index(jen.Lit(0))).Op("==").Lit(1).Block(
				jen.Id("m").Dot("handlers").Op("=").Id("m").Dot("handlers").Index(jen.Lit(1), jen.Empty()),
				jen.Id("m").Dot("lifetimes").Op("=").Id("m").Dot("lifetimes").Index(jen.Lit(1), jen.Empty()),
			).Else().If(jen.Id("m").Dot("lifetimes").Index(jen.Lit(0)).Op(">").Lit(1)).Block(
				jen.Id("m").Dot("lifetimes").Index(jen.Lit(0)).Op("--"),
			)

			if hasOutput := len(f.funcType.Outputs) > 0; hasOutput {
				g.ListFunc(f.generateOutputListWithoutName).Op(":=").Id("handler").CallFunc(f.generateInputListWithEllipsis)
			} else {
				g.Id("handler").CallFunc(f.generateInputListWithEllipsis)
			}

			g.Id("input").Op(":=").Add(f.generateInputStruct()).ValuesFunc(f.generateInputListWithoutEllipsis)
			g.Id("output").Op(":=").Add(f.generateOutputStruct()).ValuesFunc(f.generateOutputListWithoutName)
			g.Id("invoc").Op(":=").Id(f.getInvocationStructName()).Values(jen.Id("input"), jen.Id("output"))
			g.Id("m").Dot("invocations").Op("=").Append(jen.Id("m").Dot("invocations"), jen.Id("invoc"))

			g.ReturnFunc(f.generateOutputListWithoutName)
		})
}

func (f *funcMockerGeneratorHelper) noHandler() string {
	return fmt.Sprintf("%s: no handler", f.getMockerStructName())
}

func (f *funcMockerGeneratorHelper) generateInputListWithEllipsis(g *jen.Group) {
	for i, field := range f.funcType.Inputs {
		if f.funcType.IsVariadic && i == len(f.funcType.Inputs)-1 {
			g.Id(field.Name).Op("...")
		} else {
			g.Id(field.Name)
		}
	}
}

func (f *funcMockerGeneratorHelper) generateInputListWithoutEllipsis(g *jen.Group) {
	for _, field := range f.funcType.Inputs {
		g.Id(field.Name)
	}
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
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerStructName())).
		Id("TakeOneInvocation").
		Params().
		Params(jen.Id(f.getInvocationStructName())).
		BlockFunc(func(g *jen.Group) {
			f.generateLockUnlock(g)

			g.If(jen.Len(jen.Id("m").Dot("invocations")).Op("==").Lit(0)).
				Block(jen.Panic(jen.Lit(f.noInvocationPanicMessage())))

			g.Id("invoc").Op(":=").Id("m").Dot("invocations").Index(jen.Lit(0))
			g.Id("m").Dot("invocations").Op("=").Id("m").Dot("invocations").Index(jen.Lit(1), jen.Empty())
			g.Return(jen.Id("invoc"))
		})
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
