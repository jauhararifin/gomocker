package gomocker

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/dave/jennifer/jen"
)

type Namer interface {
	MockerName(typeName string) string
	ConstructorName(typeName string) string
	ArgumentName(i int, name string, needPublic bool, isReturnArgument bool) string
}

type defaultNamer struct{}

func (*defaultNamer) MockerName(typeName string) string {
	return typeName + "Mocker"
}

func (*defaultNamer) ConstructorName(typeName string) string {
	return "NewMocked" + typeName
}

func (*defaultNamer) ArgumentName(i int, name string, needPublic bool, isReturnArgument bool) string {
	if name != "" {
		if needPublic {
			return strings.ToUpper(name[:1]) + name[1:]
		}
		return name
	} else if isReturnArgument {
		if needPublic {
			return fmt.Sprintf("Out%d", i+1)
		}
		return fmt.Sprintf("out%d", i+1)
	} else if needPublic {
		return fmt.Sprintf("Arg%d", i+1)
	}
	return fmt.Sprintf("arg%d", i+1)
}

func getDeclarationByName(f *ast.File, name string) *ast.TypeSpec {
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if typeSpec.Name.String() != name {
				continue
			}

			_, IsInterface := typeSpec.Type.(*ast.InterfaceType)
			_, IsFunction := typeSpec.Type.(*ast.FuncType)
			if IsInterface || IsFunction {
				return typeSpec
			}
		}
	}
	return nil
}

type funcMockerGenerator struct {
	spec        *ast.TypeSpec
	mockerNamer Namer
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
		Add(f.generateTakeOneInvocationMethod()).Line().Line().
		Add(f.generateFuncMockerConstructor()).Line().Line()
	return code
}

func (f *funcMockerGenerator) generateMockerStruct() jen.Code {
	return jen.Type().Id(f.getMockerName()).Struct(
		jen.Id("mux").Qual("sync", "Mutex"),
		jen.Id("handlers").Index().Add(generateCodeFromExpr(f.spec.Type)),
		jen.Id("lifetimes").Index().Int(),
		jen.Id("invocations").Index().Add(f.generateInvocationStructType()),
	)
}

func (f *funcMockerGenerator) getMockerName() string {
	return f.mockerNamer.MockerName(f.spec.Name.String())
}

func (f *funcMockerGenerator) generateInvocationStructType() jen.Code {
	return jen.Struct(
		jen.Id("Inputs").Add(f.generateInputStruct()),
		jen.Id("Outputs").Add(f.generateOutputStruct()),
	)
}

func (f *funcMockerGenerator) generateInputStruct() jen.Code {
	return f.generateInputOutputStruct(f.spec.Type.(*ast.FuncType).Params, false)
}

func (f *funcMockerGenerator) generateInputOutputStruct(fieldList *ast.FieldList, isReturnParam bool) jen.Code {
	paramList := make([]jen.Code, 0, fieldList.NumFields())
	for _, paramName := range f.generateParamNames(fieldList, true, isReturnParam) {
		paramList = append(paramList, jen.Id(paramName.name).Add(generateCodeFromExpr(paramName.expr)))
	}
	return jen.Struct(paramList...)
}

type paramName struct {
	name string
	expr ast.Expr
}

func (f *funcMockerGenerator) generateParamNames(fieldList *ast.FieldList, needPublic, isReturnParam bool) []paramName {
	paramNames := make([]paramName, 0, fieldList.NumFields())

	argumentPos := 0
	for _, field := range fieldList.List {
		names := make([]string, 0, 0)
		if len(field.Names) == 0 {
			name := f.mockerNamer.ArgumentName(argumentPos, "", needPublic, isReturnParam)
			argumentPos++
			names = append(names, name)
		} else {
			for _, ident := range field.Names {
				name := f.mockerNamer.ArgumentName(argumentPos, ident.Name, needPublic, isReturnParam)
				argumentPos++
				names = append(names, name)
			}
		}

		for _, name := range names {
			paramNames = append(paramNames, paramName{
				name: name,
				expr: field.Type,
			})
		}
	}
	return paramNames
}

func (f *funcMockerGenerator) generateOutputStruct() jen.Code {
	return f.generateInputOutputStruct(f.spec.Type.(*ast.FuncType).Results, true)
}

func (f *funcMockerGenerator) generateMockMethod() jen.Code {
	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("Mock").
		Params(
			jen.Id("nTimes").Int(),
			jen.Id("f").Add(generateCodeFromExpr(f.spec.Type)),
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
	return fmt.Sprintf("%s: already mocked forever", f.getMockerName())
}

func (f *funcMockerGenerator) invalidLifetimePanicMessage() string {
	return fmt.Sprintf("%s: invalid lifetime, valid lifetime are positive number and 0 (0 means forever)", f.getMockerName())
}

func (f *funcMockerGenerator) generateMockOnceMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockOnce").
		Params(jen.Id("f").Add(generateCodeFromExpr(f.spec.Type))).
		Block(
			jen.Id("m").Dot("Mock").Call(jen.Lit(1), jen.Id("f")),
		)
}

func (f *funcMockerGenerator) generateMockForeverMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockForever").
		Params(jen.Id("f").Add(generateCodeFromExpr(f.spec.Type))).
		Block(
			jen.Id("m").Dot("Mock").Call(jen.Lit(0), jen.Id("f")),
		)
}

func (f *funcMockerGenerator) generateMockOutputsMethod() jen.Code {
	_, outputs := f.generateParamSignature(true)
	_, _, outputList := f.generateParamList()
	innerInput, innerOutput := f.generateParamSignature(false)
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

func (f *funcMockerGenerator) generateParamSignature(withOutputName bool) (inputs, outputs []jen.Code) {
	nIn := f.spec.Type.(*ast.FuncType).Params.NumFields()
	inputs = make([]jen.Code, 0, nIn)
	for _, paramName := range f.generateParamNames(f.spec.Type.(*ast.FuncType).Params, false, false) {
		inputs = append(inputs, jen.Id(paramName.name).Add(generateCodeFromExpr(paramName.expr)))
	}

	nOut := f.spec.Type.(*ast.FuncType).Results.NumFields()
	outputs = make([]jen.Code, 0, nOut)
	for _, paramName := range f.generateParamNames(f.spec.Type.(*ast.FuncType).Results, false, true) {
		if withOutputName {
			outputs = append(outputs, jen.Id(paramName.name).Add(generateCodeFromExpr(paramName.expr)))
		} else {
			outputs = append(outputs, generateCodeFromExpr(paramName.expr))
		}
	}

	return inputs, outputs
}

func (f *funcMockerGenerator) generateParamList() (inputs, inputsForCall, outputs []jen.Code) {
	inputNames := f.generateParamNames(f.spec.Type.(*ast.FuncType).Params, false, false)
	nIn := f.spec.Type.(*ast.FuncType).Params.NumFields()
	inputsForCall = make([]jen.Code, 0, nIn)
	inputs = make([]jen.Code, 0, nIn)
	for _, inputName := range inputNames {
		callParam := jen.Id(inputName.name)
		if _, ok := inputName.expr.(*ast.Ellipsis); ok {
			callParam.Op("...")
		}
		inputsForCall = append(inputsForCall, callParam)
		inputs = append(inputs, jen.Id(inputName.name))
	}

	outputNames := f.generateParamNames(f.spec.Type.(*ast.FuncType).Results, false, true)
	nOut := f.spec.Type.(*ast.FuncType).Results.NumFields()
	outputs = make([]jen.Code, 0, nOut)
	for _, outputName := range outputNames {
		outputs = append(outputs, jen.Id(outputName.name))
	}

	return
}

func (f *funcMockerGenerator) generateMockOutputsOnceMethod() jen.Code {
	_, outputs := f.generateParamSignature(true)
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
	_, outputs := f.generateParamSignature(true)
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

	outputNames := f.generateParamNames(f.spec.Type.(*ast.FuncType).Results, false, true)
	for _, outputName := range outputNames {
		body = append(
			body,
			jen.Var().Add(jen.Id(outputName.name)).Add(generateCodeFromExpr(outputName.expr)),
		)
	}

	mockOutputParams := []jen.Code{jen.Id("nTimes")}
	for _, outputName := range outputNames {
		mockOutputParams = append(mockOutputParams, jen.Id(outputName.name))
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
	params, returns := f.generateParamSignature(true)
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
	return fmt.Sprintf("%s: no handler", f.getMockerName())
}

func (f *funcMockerGenerator) generateInvocationsMethod() jen.Code {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("Invocations").
		Params().
		Params(jen.Index().Add(f.generateInvocationStructType())).
		Block(jen.Return(jen.Id("m").Dot("invocations")))
}

func (f *funcMockerGenerator) generateTakeOneInvocationMethod() jen.Code {
	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
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
	return fmt.Sprintf("%s: no invocations", f.getMockerName())
}

func (f *funcMockerGenerator) generateFuncMockerConstructor() jen.Code {
	return jen.Func().
		Id(f.getMockerConstructorName()).
		Params().
		Params(
			generateCodeFromExpr(f.spec.Type),
			jen.Op("*").Id(f.getMockerName()),
		).
		Block(
			jen.Id("m").Op(":=").Op("&").Id(f.getMockerName()).Values(),
			jen.Return(jen.Id("m").Dot("Call"), jen.Id("m")),
		)
}

func (f *funcMockerGenerator) getMockerConstructorName() string {
	return f.mockerNamer.ConstructorName(f.spec.Name.String())
}
