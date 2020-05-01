package gomocker

import (
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"strings"

	"github.com/dave/jennifer/jen"
)

type exprCodeGenerator struct {
	importMap map[string]string
}

func (f *exprCodeGenerator) generateCodeFromExpr(e ast.Expr) (jen.Code, error) {
	switch v := e.(type) {
	case *ast.UnaryExpr:
		return f.generateCodeFromUnaryExpr(v)
	case *ast.BinaryExpr:
		return f.generateCodeFromBinaryExpr(v)
	case *ast.BasicLit:
		return f.generateCodeFromBasicLit(v)
	case *ast.SelectorExpr:
		return f.generateCodeFromSelectorExpr(v)
	case *ast.Ident:
		return f.generateCodeFromIdent(v)
	case *ast.Ellipsis:
		return f.generateCodeFromEllipsis(v)
	case *ast.StarExpr:
		return f.generateCodeFromStarExpr(v)
	case *ast.ArrayType:
		return f.generateCodeFromArrayType(v)
	case *ast.FuncType:
		return f.generateCodeFromFuncType(v)
	case *ast.MapType:
		return f.generateCodeFromMapType(v)
	case *ast.ChanType:
		return f.generateCodeFromChanType(v)
	case *ast.StructType:
		return f.generateCodeFromStructType(v)
	case *ast.InterfaceType:
		return f.generateCodeFromInterfaceType(v)
	}
	return nil, fmt.Errorf("found unrecognized type: %v", e)
}

func (f *exprCodeGenerator) generateCodeFromUnaryExpr(unaryExpr *ast.UnaryExpr) (jen.Code, error) {
	x, err := f.generateCodeFromExpr(unaryExpr.X)
	if err != nil {
		return nil, err
	}
	return jen.Op(unaryExpr.Op.String()).Add(x), nil
}

func (f *exprCodeGenerator) generateCodeFromBinaryExpr(binaryExpr *ast.BinaryExpr) (jen.Code, error) {
	x, err := f.generateCodeFromExpr(binaryExpr.X)
	if err != nil {
		return nil, err
	}
	y, err := f.generateCodeFromExpr(binaryExpr.Y)
	if err != nil {
		return nil, err
	}
	return jen.Add(x).Op(binaryExpr.Op.String()).Add(y), nil
}

func (f *exprCodeGenerator) generateCodeFromBasicLit(basicLit *ast.BasicLit) (jen.Code, error) {
	switch basicLit.Kind {
	case token.INT:
		i, ok := parseInt(basicLit.Value)
		if !ok {
			return nil, fmt.Errorf("invalid_int: %s", basicLit.Value)
		}
		return jen.Lit(i), nil
	case token.FLOAT:
		return nil, fmt.Errorf("not_supported_yet: cannot_convert_float_literal: %v", basicLit)
	case token.CHAR:
		return nil, fmt.Errorf("not_supported_yet: cannot_convert_char_literal: %v", basicLit)
	case token.STRING:
		return nil, fmt.Errorf("not_supported_yet: cannot_convert_string_literal: %v", basicLit)
	case token.IMAG:
		return nil, fmt.Errorf("not_supported_yet: cannot_convert_imag_literal: %v", basicLit)
	}
	return nil, fmt.Errorf("cannot_parse_basic_lit: %v", basicLit)
}

func (f *exprCodeGenerator) generateCodeFromIdent(ident *ast.Ident) (jen.Code, error) {
	switch ident.Name {
	case "bool":
		return jen.Bool(), nil
	case "int":
		return jen.Int(), nil
	case "int8":
		return jen.Int8(), nil
	case "int16":
		return jen.Int16(), nil
	case "int32":
		return jen.Int32(), nil
	case "int64":
		return jen.Int64(), nil
	case "uint":
		return jen.Uint(), nil
	case "uint8":
		return jen.Uint8(), nil
	case "uint16":
		return jen.Uint16(), nil
	case "uint32":
		return jen.Uint32(), nil
	case "uint64":
		return jen.Uint64(), nil
	case "uintptr":
		return jen.Uintptr(), nil
	case "float32":
		return jen.Float32(), nil
	case "float64":
		return jen.Float64(), nil
	case "complex64":
		return jen.Complex64(), nil
	case "complex128":
		return jen.Complex128(), nil
	case "string":
		return jen.String(), nil
	case "error":
		return jen.Error(), nil
	}
	return jen.Id(ident.Name), nil
}

func (f *exprCodeGenerator) generateCodeFromSelectorExpr(selectorExpr *ast.SelectorExpr) (jen.Code, error) {
	if ident, ok := selectorExpr.X.(*ast.Ident); ok {
		importPath, ok := f.importMap[ident.String()]
		if !ok {
			return nil, fmt.Errorf("unrecognized identifier: %s", ident.String())
		}
		return jen.Qual(importPath, selectorExpr.Sel.String()), nil
	}

	x, err := f.generateCodeFromExpr(selectorExpr.X)
	if err != nil {
		return nil, err
	}
	return jen.Add(x).Dot(selectorExpr.Sel.String()), nil
}

func (f *exprCodeGenerator) generateCodeFromEllipsis(ellipsis *ast.Ellipsis) (jen.Code, error) {
	elt, err := f.generateCodeFromExpr(ellipsis.Elt)
	if err != nil {
		return nil, err
	}
	return jen.Op("...").Add(elt), nil
}

func (f *exprCodeGenerator) generateCodeFromStarExpr(starExpr *ast.StarExpr) (jen.Code, error) {
	x, err := f.generateCodeFromExpr(starExpr.X)
	if err != nil {
		return nil, err
	}
	return jen.Op("*").Add(x), nil
}

func (f *exprCodeGenerator) generateCodeFromArrayType(arrayType *ast.ArrayType) (jen.Code, error) {
	elt, err := f.generateCodeFromExpr(arrayType.Elt)
	if err != nil {
		return nil, err
	}

	if arrayType.Len == nil {
		return jen.Index().Add(elt), nil
	}

	lenn, err := f.generateCodeFromExpr(arrayType.Len)
	if err != nil {
		return nil, err
	}
	return jen.Index(lenn).Add(elt), nil
}

func (f *exprCodeGenerator) generateCodeFromFuncType(funcType *ast.FuncType) (jen.Code, error) {
	params, err := f.generateCodeFromFieldList(funcType.Params)
	if err != nil {
		return nil, err
	}

	results, err := f.generateCodeFromFieldList(funcType.Results)
	if err != nil {
		return nil, err
	}

	return jen.Func().Params(params...).Params(results...), nil
}

func (f *exprCodeGenerator) generateCodeFromFieldList(fields *ast.FieldList) ([]jen.Code, error) {
	params := make([]jen.Code, 0, fields.NumFields())
	for _, field := range fields.List {
		typeCode, err := f.generateCodeFromExpr(field.Type)
		if err != nil {
			return nil, err
		}

		if len(field.Names) == 0 {
			params = append(params, typeCode)
			continue
		}

		for _, id := range field.Names {
			params = append(params, jen.Id(id.Name).Add(typeCode))
		}
	}
	return params, nil
}

func (f *exprCodeGenerator) generateCodeFromMapType(mapType *ast.MapType) (jen.Code, error) {
	keyDef, err := f.generateCodeFromExpr(mapType.Key)
	if err != nil {
		return nil, err
	}

	valDef, err := f.generateCodeFromExpr(mapType.Value)
	if err != nil {
		return nil, err
	}

	return jen.Map(keyDef).Add(valDef), nil
}

func (f *exprCodeGenerator) generateCodeFromChanType(chanType *ast.ChanType) (jen.Code, error) {
	c, err := f.generateCodeFromExpr(chanType.Value)
	if err != nil {
		return nil, err
	}

	switch chanType.Dir {
	case ast.RECV:
		return jen.Op("<-").Chan().Add(c), nil
	case ast.SEND:
		return jen.Chan().Op("<-").Add(c), nil
	default:
		return jen.Chan().Add(c), nil
	}
}

func (f *exprCodeGenerator) generateCodeFromStructType(structType *ast.StructType) (jen.Code, error) {
	params := make([]jen.Code, 0, structType.Fields.NumFields())
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			typ, err := f.generateCodeFromExpr(field.Type)
			if err != nil {
				return nil, err
			}
			params = append(
				params,
				jen.Id(name.String()).Add(typ),
			)
		}
	}
	return jen.Struct(params...), nil
}

func (f *exprCodeGenerator) generateCodeFromInterfaceType(interfaceType *ast.InterfaceType) (jen.Code, error) {
	nMethod := interfaceType.Methods.NumFields()
	methods := make([]jen.Code, 0, nMethod)
	for _, field := range interfaceType.Methods.List {
		name := field.Names[0].String()
		funcType := field.Type.(*ast.FuncType)

		params, err := f.generateCodeFromFieldList(funcType.Params)
		if err != nil {
			return nil, err
		}

		results, err := f.generateCodeFromFieldList(funcType.Results)
		if err != nil {
			return nil, err
		}

		methods = append(methods, jen.Id(name).Params(params...).Params(results...))
	}
	return jen.Interface(methods...), nil
}

func newExprCodeGenerator(f *ast.File) *exprCodeGenerator {
	importMap := make(map[string]string)
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}

			importPath := importSpec.Path.Value
			importPath = strings.TrimPrefix(strings.TrimSuffix(importPath, "\""), "\"")

			importName := path.Base(importPath)
			if importSpec.Name != nil {
				importName = importSpec.Name.String()
			}

			importMap[importName] = importPath
		}
	}

	return &exprCodeGenerator{importMap: importMap}
}
