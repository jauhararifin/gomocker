package gomocker

import (
	"fmt"
	"go/ast"
	"path"
	"strings"
)

func TypeFromAstName(file *ast.File, name string) Type {
	return (&astTypeGenerator{}).generateTypeFromAst(file, name)
}

type astTypeGenerator struct {
	importMap map[string]string
}

func (f *astTypeGenerator) generateTypeFromAst(file *ast.File, name string) Type {
	spec := f.getDeclarationByName(file, name)
	if spec == nil {
		panic(fmt.Errorf("definition not found: %v", name))
	}

	exprCodeGen := newExprCodeGenerator(file)
	return exprCodeGen.generateTypeFromExpr(spec.Type)
}

func (*astTypeGenerator) getDeclarationByName(f *ast.File, name string) *ast.TypeSpec {
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

func (f *astTypeGenerator) generateTypeFromExpr(e ast.Expr) Type {
	switch v := e.(type) {
	case *ast.SelectorExpr:
		return f.generateTypeFromSelectorExpr(v)
	case *ast.Ident:
		return f.generateTypeFromIdent(v)
	case *ast.StarExpr:
		typ := f.generateTypeFromStarExpr(v)
		return Type{PtrType: &typ}
	case *ast.ArrayType:
		return f.generateTypeFromArrayType(v)
	case *ast.FuncType:
		typ := f.generateTypeFromFuncType(v)
		return Type{FuncType: &typ}
	case *ast.MapType:
		typ := f.generateTypeFromMapType(v)
		return Type{MapType: &typ}
	case *ast.ChanType:
		typ := f.generateTypeFromChanType(v)
		return Type{ChanType: &typ}
	case *ast.StructType:
		typ := f.generateTypeFromStructType(v)
		return Type{StructType: &typ}
	case *ast.InterfaceType:
		typ := f.generateTypeFromInterfaceType(v)
		return Type{InterfaceType: &typ}
	}
	panic(fmt.Errorf("unrecognized type: %v", e))
}

func (f *astTypeGenerator) generateTypeFromIdent(ident *ast.Ident) Type {
	switch ident.Name {
	case string(PrimitiveKindBool):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindBool}}
	case string(PrimitiveKindInt):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindInt}}
	case string(PrimitiveKindInt8):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindInt8}}
	case string(PrimitiveKindInt16):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindInt16}}
	case string(PrimitiveKindInt32):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindInt32}}
	case string(PrimitiveKindInt64):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindInt64}}
	case string(PrimitiveKindUint):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindUint}}
	case string(PrimitiveKindUint8):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindUint8}}
	case string(PrimitiveKindUint16):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindUint16}}
	case string(PrimitiveKindUint32):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindUint32}}
	case string(PrimitiveKindUint64):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindUint64}}
	case string(PrimitiveKindUintptr):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindUintptr}}
	case string(PrimitiveKindFloat32):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindFloat32}}
	case string(PrimitiveKindFloat64):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindFloat64}}
	case string(PrimitiveKindComplex64):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindComplex64}}
	case string(PrimitiveKindComplex128):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindComplex128}}
	case string(PrimitiveKindString):
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindString}}
	case "error":
		return Type{PrimitiveType: &PrimitiveType{Kind: PrimitiveKindError}}
	}
	panic(fmt.Errorf("unrecognized type: %v", ident.Name))
}

func (f *astTypeGenerator) generateTypeFromSelectorExpr(selectorExpr *ast.SelectorExpr) Type {
	ident, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		panic(fmt.Errorf("unrecognized expr: %v", selectorExpr))
	}

	importPath, ok := f.importMap[ident.String()]
	if !ok {
		panic(fmt.Errorf("unrecognized identifier: %s", ident.String()))
	}
	return Type{QualType: &QualType{
		Package: importPath,
		Name:    selectorExpr.Sel.String(),
	}}
}

func (f *astTypeGenerator) generateTypeFromStarExpr(starExpr *ast.StarExpr) PtrType {
	return PtrType{Elem: f.generateTypeFromExpr(starExpr.X)}
}

func (f *astTypeGenerator) generateTypeFromArrayType(arrayType *ast.ArrayType) Type {
	if arrayType.Len == nil {
		return Type{SliceType: &SliceType{Elem: f.generateTypeFromExpr(arrayType.Elt)}}
	}

	lit, ok := arrayType.Len.(*ast.BasicLit)
	if !ok {
		panic(fmt.Errorf("unrecognized array length: %v", arrayType.Len))
	}
	lenn, ok := parseInt(lit.Value)
	if !ok {
		panic(fmt.Errorf("unrecognized array length: %v", lit.Value))
	}

	return Type{ArrayType: &ArrayType{
		Len:  lenn,
		Elem: f.generateTypeFromExpr(arrayType.Elt),
	}}
}

func (f *astTypeGenerator) generateTypeFromFuncType(funcType *ast.FuncType) FuncType {
	params, isVariadic := f.generateTypeFromFieldList(funcType.Params, f.getInputNamesFromAst(funcType.Params.List))

	var results []TypeField = nil
	if funcType.Results != nil {
		results, _ = f.generateTypeFromFieldList(funcType.Results, f.getOutputNamesFromAst(funcType.Results.List))
	}
	return FuncType{
		Inputs:     params,
		Outputs:    results,
		IsVariadic: isVariadic,
	}
}

func (f *astTypeGenerator) generateTypeFromFieldList(fields *ast.FieldList, names []string) (types []TypeField, isVariadic bool) {
	if fields == nil {
		return nil, false
	}

	types = make([]TypeField, 0, fields.NumFields())
	i := 0
	for _, field := range fields.List {
		typExpr := field.Type
		if v, ok := field.Type.(*ast.Ellipsis); ok {
			isVariadic = true
			typExpr = v.Elt
		}
		typ := f.generateTypeFromExpr(typExpr)

		if len(field.Names) == 0 {
			types = append(types, TypeField{
				Name: names[i],
				Type: typ,
			})
			i++
			continue
		}

		for range field.Names {
			types = append(types, TypeField{
				Name: names[i],
				Type: typ,
			})
			i++
		}
	}

	return
}

func (f *astTypeGenerator) getInputNamesFromAst(inputs []*ast.Field) []string {
	return f.getNamesFromExpr(inputs, "arg")
}

func (f *astTypeGenerator) getOutputNamesFromAst(inputs []*ast.Field) []string {
	return f.getNamesFromExpr(inputs, "out")
}

func (f *astTypeGenerator) getNamesFromExpr(params []*ast.Field, prefix string) []string {
	names := make([]string, 0, len(params))
	i := 0
	for _, p := range params {
		n := len(p.Names)
		if n == 0 {
			n = 1
		}
		for j := 0; j < n; j++ {
			if len(p.Names) > 0 {
				names = append(names, p.Names[j].String())
			} else {
				i++
				names = append(names, fmt.Sprintf("%s%d", prefix, i))
			}
		}
	}
	return names
}

func (f *astTypeGenerator) generateTypeFromMapType(mapType *ast.MapType) MapType {
	keyDef := f.generateTypeFromExpr(mapType.Key)
	valDef := f.generateTypeFromExpr(mapType.Value)
	return MapType{
		Key:  keyDef,
		Elem: valDef,
	}
}

func (f *astTypeGenerator) generateTypeFromChanType(chanType *ast.ChanType) ChanType {
	c := f.generateTypeFromExpr(chanType.Value)

	switch chanType.Dir {
	case ast.RECV:
		return ChanType{Dir: ChanTypeDirRecv, Elem: c}
	case ast.SEND:
		return ChanType{Dir: ChanTypeDirSend, Elem: c}
	default:
		return ChanType{Dir: ChanTypeDirBoth, Elem: c}
	}
}

func (f *astTypeGenerator) generateTypeFromStructType(structType *ast.StructType) StructType {
	if structType.Fields == nil {
		return StructType{Fields: nil}
	}

	fields := make([]TypeField, 0, structType.Fields.NumFields())
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			fields = append(
				fields,
				TypeField{
					Name: name.String(),
					Type: f.generateTypeFromExpr(field.Type),
				},
			)
		}
	}

	return StructType{Fields: fields}
}

func (f *astTypeGenerator) generateTypeFromInterfaceType(interfaceType *ast.InterfaceType) InterfaceType {
	if interfaceType.Methods == nil {
		return InterfaceType{Methods: nil}
	}

	nMethod := interfaceType.Methods.NumFields()
	methods := make([]InterfaceTypeMethod, 0, nMethod)
	for _, field := range interfaceType.Methods.List {
		name := field.Names[0].String()
		funcType := field.Type.(*ast.FuncType)
		methods = append(methods, InterfaceTypeMethod{
			Name: name,
			Func: f.generateTypeFromFuncType(funcType),
		})
	}

	return InterfaceType{Methods: methods}
}

func newExprCodeGenerator(f *ast.File) *astTypeGenerator {
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

	return &astTypeGenerator{importMap: importMap}
}
