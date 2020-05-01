package gomocker

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/dave/jennifer/jen"
)

func generateCodeFromExpr(e ast.Expr) (jen.Code, error) {
	switch v := e.(type) {
	case *ast.UnaryExpr:
		return generateCodeFromUnaryExpr(v)
	case *ast.BinaryExpr:
		return generateCodeFromBinaryExpr(v)
	case *ast.BasicLit:
		return generateCodeFromBasicLit(v)
	case *ast.SelectorExpr:
		return generateCodeFromSelectorExpr(v)
	case *ast.Ident:
		return generateCodeFromIdent(v)
	case *ast.Ellipsis:
		return generateCodeFromEllipsis(v)
	case *ast.StarExpr:
		return generateCodeFromStarExpr(v)
	case *ast.ArrayType:
		return generateCodeFromArrayType(v)
	case *ast.FuncType:
		return generateCodeFromFuncType(v)
	case *ast.MapType:
		return generateCodeFromMapType(v)
	case *ast.ChanType:
		return generateCodeFromChanType(v)
	case *ast.StructType:
		return generateCodeFromStructType(v)
	case *ast.InterfaceType:
		return generateCodeFromInterfaceType(v)
	}
	return nil, fmt.Errorf("found unrecognized type: %v", e)
}

func generateCodeFromUnaryExpr(unaryExpr *ast.UnaryExpr) (jen.Code, error) {
	x, err := generateCodeFromExpr(unaryExpr.X)
	if err != nil {
		return nil, err
	}
	return jen.Op(unaryExpr.Op.String()).Add(x), nil
}

func generateCodeFromBinaryExpr(binaryExpr *ast.BinaryExpr) (jen.Code, error) {
	x, err := generateCodeFromExpr(binaryExpr.X)
	if err != nil {
		return nil, err
	}
	y, err := generateCodeFromExpr(binaryExpr.Y)
	if err != nil {
		return nil, err
	}
	return jen.Add(x).Op(binaryExpr.Op.String()).Add(y), nil
}

func generateCodeFromBasicLit(basicLit *ast.BasicLit) (jen.Code, error) {
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

func parseInt(s string) (int, bool) {
	if len(s) == 0 {
		return 0, false
	}

	if len(s) == 1 {
		if s[0] < '0' || s[0] > '9' {
			return 0, false
		}
		return int(s[0]) - '0', true
	}

	for i := 1; i < len(s); i++ {
		if s[i] == '_' && s[i-1] == '_' {
			return 0, false
		}
	}
	if s[len(s)-1] == '_' || s[0] == '_' {
		return 0, false
	}

	if s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		return parseIntHex(s[2:])
	} else if s[0] == '0' && (s[1] == 'b' || s[1] == 'B') {
		return parseIntBin(s[2:])
	} else if s[0] == '0' {
		if s[1] == 'o' || s[1] == 'O' {
			return parseIntOct(s[2:])
		}
		return parseIntOct(s[1:])
	}

	s = strings.ReplaceAll(s, "_", "")
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + int(s[i]) - '0'
	}
	return n, true
}

func parseIntHex(s string) (int, bool) {
	s = strings.ReplaceAll(s, "_", "")
	n := 0
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] >= 'a' && s[i] <= 'f':
			n = n*16 + int(s[i]) - 'a' + 10
		case s[i] >= 'A' && s[i] <= 'F':
			n = n*16 + int(s[i]) - 'A' + 10
		case s[i] >= '0' && s[i] <= '9':
			n = n*16 + int(s[i]) - '0'
		}
	}
	return n, true
}

func parseIntOct(s string) (int, bool) {
	s = strings.ReplaceAll(s, "_", "")
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '7' {
			return 0, false
		}
		n = n*8 + int(s[i]) - '0'
	}
	return n, true
}

func parseIntBin(s string) (int, bool) {
	s = strings.ReplaceAll(s, "_", "")
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] != '0' && s[i] != '1' {
			return 0, false
		}
		n = n*2 + int(s[i]) - '0'
	}
	return n, true
}

func generateCodeFromIdent(ident *ast.Ident) (jen.Code, error) {
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
	}
	return jen.Id(ident.Name), nil
}

func generateCodeFromSelectorExpr(selectorExpr *ast.SelectorExpr) (jen.Code, error) {
	x, err := generateCodeFromExpr(selectorExpr.X)
	if err != nil {
		return nil, err
	}
	return jen.Add(x).Dot(selectorExpr.Sel.Name), nil
}

func generateCodeFromEllipsis(ellipsis *ast.Ellipsis) (jen.Code, error) {
	elt, err := generateCodeFromExpr(ellipsis.Elt)
	if err != nil {
		return nil, err
	}
	return jen.Op("...").Add(elt), nil
}

func generateCodeFromStarExpr(starExpr *ast.StarExpr) (jen.Code, error) {
	x, err := generateCodeFromExpr(starExpr.X)
	if err != nil {
		return nil, err
	}
	return jen.Op("*").Add(x), nil
}

func generateCodeFromArrayType(arrayType *ast.ArrayType) (jen.Code, error) {
	elt, err := generateCodeFromExpr(arrayType.Elt)
	if err != nil {
		return nil, err
	}

	if arrayType.Len == nil {
		return jen.Index().Add(elt), nil
	}

	lenn, err := generateCodeFromExpr(arrayType.Len)
	if err != nil {
		return nil, err
	}
	return jen.Index(lenn).Add(elt), nil
}

func generateCodeFromFuncType(funcType *ast.FuncType) (jen.Code, error) {
	params, err := generateCodeFromFieldList(funcType.Params)
	if err != nil {
		return nil, err
	}

	results, err := generateCodeFromFieldList(funcType.Results)
	if err != nil {
		return nil, err
	}

	return jen.Func().Params(params...).Params(results...), nil
}

func generateCodeFromFieldList(fields *ast.FieldList) ([]jen.Code, error) {
	params := make([]jen.Code, 0, fields.NumFields())
	for _, field := range fields.List {
		typeCode, err := generateCodeFromExpr(field.Type)
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

func generateCodeFromMapType(mapType *ast.MapType) (jen.Code, error) {
	keyDef, err := generateCodeFromExpr(mapType.Key)
	if err != nil {
		return nil, err
	}

	valDef, err := generateCodeFromExpr(mapType.Value)
	if err != nil {
		return nil, err
	}

	return jen.Map(keyDef).Add(valDef), nil
}

func generateCodeFromChanType(chanType *ast.ChanType) (jen.Code, error) {
	c, err := generateCodeFromExpr(chanType.Value)
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

func generateCodeFromStructType(structType *ast.StructType) (jen.Code, error) {
	params := make([]jen.Code, 0, structType.Fields.NumFields())
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			typ, err := generateCodeFromExpr(field.Type)
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

func generateCodeFromInterfaceType(interfaceType *ast.InterfaceType) (jen.Code, error) {
	nMethod := interfaceType.Methods.NumFields()
	methods := make([]jen.Code, 0, nMethod)
	for _, field := range interfaceType.Methods.List {
		name := field.Names[0].String()
		funcType := field.Type.(*ast.FuncType)

		params, err := generateCodeFromFieldList(funcType.Params)
		if err != nil {
			return nil, err
		}

		results, err := generateCodeFromFieldList(funcType.Results)
		if err != nil {
			return nil, err
		}

		methods = append(methods, jen.Id(name).Params(params...).Params(results...))
	}
	return jen.Interface(methods...), nil
}
