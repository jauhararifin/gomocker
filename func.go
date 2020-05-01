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

func (f *funcMockerGenerator) generate() (jen.Code, error) {
	return f.concatSteps(
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
		f.generateFuncMockerConstructor,
	)
}

func (f *funcMockerGenerator) concatSteps(steps ...func() (jen.Code, error)) (jen.Code, error) {
	j := jen.Empty()
	for _, step := range steps {
		code, err := step()
		if err != nil {
			return nil, err
		}
		j.Add(code).Line().Line()
	}
	return j, nil
}

func (f *funcMockerGenerator) generateMockerStruct() (jen.Code, error) {
	typ, err := generateCodeFromExpr(f.spec.Type)
	if err != nil {
		return nil, err
	}

	invocStruct, err := f.generateInvocationStructType()
	if err != nil {
		return nil, err
	}

	return jen.Type().Id(f.getMockerName()).Struct(
		jen.Id("mux").Qual("sync", "Mutex"),
		jen.Id("handlers").Index().Add(typ),
		jen.Id("lifetimes").Index().Int(),
		jen.Id("invocations").Index().Add(invocStruct),
	), nil
}

func (f *funcMockerGenerator) getMockerName() string {
	return f.mockerNamer.MockerName(f.spec.Name.String())
}

func (f *funcMockerGenerator) generateInvocationStructType() (jen.Code, error) {
	inputStruct, err := f.generateInputStruct()
	if err != nil {
		return nil, err
	}

	outputStruct, err := f.generateOutputStruct()
	if err != nil {
		return nil, err
	}

	return jen.Struct(
		jen.Id("Inputs").Add(inputStruct),
		jen.Id("Outputs").Add(outputStruct),
	), nil
}

func (f *funcMockerGenerator) generateInputStruct() (jen.Code, error) {
	return f.generateInputOutputStruct(f.spec.Type.(*ast.FuncType).Params, false)
}

func (f *funcMockerGenerator) generateInputOutputStruct(fieldList *ast.FieldList, isReturnParam bool) (jen.Code, error) {
	paramList := make([]jen.Code, 0, fieldList.NumFields())
	for _, paramName := range f.generateParamNames(fieldList, true, isReturnParam) {
		typ, err := generateCodeFromExpr(paramName.expr)
		if err != nil {
			return nil, err
		}
		paramList = append(paramList, jen.Id(paramName.name).Add(typ))
	}
	return jen.Struct(paramList...), nil
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

func (f *funcMockerGenerator) generateOutputStruct() (jen.Code, error) {
	return f.generateInputOutputStruct(f.spec.Type.(*ast.FuncType).Results, true)
}

func (f *funcMockerGenerator) generateMockMethod() (jen.Code, error) {
	typ, err := generateCodeFromExpr(f.spec.Type)
	if err != nil {
		return nil, err
	}

	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("Mock").
		Params(
			jen.Id("nTimes").Int(),
			jen.Id("f").Add(typ),
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
	return code, nil
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

func (f *funcMockerGenerator) generateMockOnceMethod() (jen.Code, error) {
	typ, err := generateCodeFromExpr(f.spec.Type)
	if err != nil {
		return nil, err
	}
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockOnce").
		Params(jen.Id("f").Add(typ)).
		Block(
			jen.Id("m").Dot("Mock").Call(jen.Lit(1), jen.Id("f")),
		), nil
}

func (f *funcMockerGenerator) generateMockForeverMethod() (jen.Code, error) {
	typ, err := generateCodeFromExpr(f.spec.Type)
	if err != nil {
		return nil, err
	}

	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockForever").
		Params(jen.Id("f").Add(typ)).
		Block(
			jen.Id("m").Dot("Mock").Call(jen.Lit(0), jen.Id("f")),
		), nil
}

func (f *funcMockerGenerator) generateMockOutputsMethod() (jen.Code, error) {
	_, outputs, err := f.generateParamSignature(true)
	if err != nil {
		return nil, err
	}

	_, _, outputList := f.generateParamList()

	innerInput, innerOutput, err := f.generateParamSignature(false)
	if err != nil {
		return nil, err
	}

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
	), nil
}

func (f *funcMockerGenerator) generateParamSignature(withOutputName bool) (inputs, outputs []jen.Code, err error) {
	nIn := f.spec.Type.(*ast.FuncType).Params.NumFields()
	inputs = make([]jen.Code, 0, nIn)
	for _, paramName := range f.generateParamNames(f.spec.Type.(*ast.FuncType).Params, false, false) {
		expr, err := generateCodeFromExpr(paramName.expr)
		if err != nil {
			return nil, nil, err
		}
		inputs = append(inputs, jen.Id(paramName.name).Add(expr))
	}

	nOut := f.spec.Type.(*ast.FuncType).Results.NumFields()
	outputs = make([]jen.Code, 0, nOut)
	for _, paramName := range f.generateParamNames(f.spec.Type.(*ast.FuncType).Results, false, true) {
		expr, err := generateCodeFromExpr(paramName.expr)
		if err != nil {
			return nil, nil, err
		}

		if withOutputName {
			outputs = append(outputs, jen.Id(paramName.name).Add(expr))
		} else {
			outputs = append(outputs, expr)
		}
	}

	return inputs, outputs, nil
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

func (f *funcMockerGenerator) generateMockOutputsOnceMethod() (jen.Code, error) {
	_, outputs, err := f.generateParamSignature(true)
	if err != nil {
		return nil, err
	}

	_, _, outputList := f.generateParamList()

	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockOutputsOnce").
		Params(outputs...).
		Block(
			jen.Id("m").Dot("MockOutputs").Call(
				append([]jen.Code{jen.Lit(1)}, outputList...)...
			),
		), nil
}

func (f *funcMockerGenerator) generateMockOutputsForeverMethod() (jen.Code, error) {
	_, outputs, err := f.generateParamSignature(true)
	if err != nil {
		return nil, err
	}

	_, _, outputList := f.generateParamList()

	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockOutputsForever").
		Params(outputs...).
		Block(
			jen.Id("m").Dot("MockOutputs").Call(
				append([]jen.Code{jen.Lit(0)}, outputList...)...
			),
		), nil
}

func (f *funcMockerGenerator) generateMockDefaultsMethod() (jen.Code, error) {
	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockDefaults").
		Params(jen.Id("nTimes").Int())

	body := make([]jen.Code, 0, 0)

	outputNames := f.generateParamNames(f.spec.Type.(*ast.FuncType).Results, false, true)
	for _, outputName := range outputNames {
		expr, err := generateCodeFromExpr(outputName.expr)
		if err != nil {
			return nil, err
		}
		body = append(
			body,
			jen.Var().Add(jen.Id(outputName.name)).Add(expr),
		)
	}

	mockOutputParams := []jen.Code{jen.Id("nTimes")}
	for _, outputName := range outputNames {
		mockOutputParams = append(mockOutputParams, jen.Id(outputName.name))
	}
	body = append(body, jen.Id("m").Dot("MockOutputs").Call(mockOutputParams...))
	code.Block(body...)

	return code, nil
}

func (f *funcMockerGenerator) generateMockDefaultsOnce() (jen.Code, error) {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockDefaultsOnce").
		Params().
		Block(jen.Id("m").Dot("MockDefaults").Call(jen.Lit(1))), nil
}

func (f *funcMockerGenerator) generateMockDefaultsForever() (jen.Code, error) {
	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("MockDefaultsForever").
		Params().
		Block(jen.Id("m").Dot("MockDefaults").Call(jen.Lit(0))), nil
}

func (f *funcMockerGenerator) generateCallMethod() (jen.Code, error) {
	params, returns, err := f.generateParamSignature(true)
	if err != nil {
		return nil, err
	}

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

	inputStruct, err := f.generateInputStruct()
	if err != nil {
		return nil, err
	}

	outputStruct, err := f.generateOutputStruct()
	if err != nil {
		return nil, err
	}

	invocStruct, err := f.generateInvocationStructType()
	if err != nil {
		return nil, err
	}

	body = append(
		body,
		jen.Id("input").Op(":=").Add(inputStruct).Values(inputs...),
		jen.Id("output").Op(":=").Add(outputStruct).Values(outputs...),
		jen.Id("invoc").Op(":=").Add(invocStruct).Values(jen.Id("input"), jen.Id("output")),
		jen.Id("m").Dot("invocations").Op("=").Append(jen.Id("m").Dot("invocations"), jen.Id("invoc")),
	)

	body = append(body, jen.Return(outputs...))

	code.Block(body...)
	return code, nil
}

func (f *funcMockerGenerator) noHandler() string {
	return fmt.Sprintf("%s: no handler", f.getMockerName())
}

func (f *funcMockerGenerator) generateInvocationsMethod() (jen.Code, error) {
	invocStruct, err := f.generateInvocationStructType()
	if err != nil {
		return nil, err
	}

	return jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("Invocations").
		Params().
		Params(jen.Index().Add(invocStruct)).
		Block(jen.Return(jen.Id("m").Dot("invocations"))), nil
}

func (f *funcMockerGenerator) generateTakeOneInvocationMethod() (jen.Code, error) {
	invocStruct, err := f.generateInvocationStructType()
	if err != nil {
		return nil, err
	}

	code := jen.Func().
		Params(jen.Id("m").Op("*").Id(f.getMockerName())).
		Id("TakeOneInvocation").
		Params().
		Params(invocStruct)

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
	return code, nil
}

func (f *funcMockerGenerator) noInvocationPanicMessage() string {
	return fmt.Sprintf("%s: no invocations", f.getMockerName())
}

func (f *funcMockerGenerator) generateFuncMockerConstructor() (jen.Code, error) {
	typ, err := generateCodeFromExpr(f.spec.Type)
	if err != nil {
		return nil, err
	}

	return jen.Func().
		Id(f.getMockerConstructorName()).
		Params().
		Params(
			typ,
			jen.Op("*").Id(f.getMockerName()),
		).
		Block(
			jen.Id("m").Op(":=").Op("&").Id(f.getMockerName()).Values(),
			jen.Return(jen.Id("m").Dot("Call"), jen.Id("m")),
		), nil
}

func (f *funcMockerGenerator) getMockerConstructorName() string {
	return f.mockerNamer.ConstructorName(f.spec.Name.String())
}
