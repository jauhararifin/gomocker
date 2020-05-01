package gomocker

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/dave/jennifer/jen"
)

func generateCodeFromExpr(e ast.Expr) jen.Code {
	switch v := e.(type) {
	case *ast.UnaryExpr:
		return generateCodeFromUnaryExpr(v)
	case *ast.BinaryExpr:
		return generateCodeFromBinaryExpr(v)
	case *ast.BasicLit:
		return generateCodeFromBasicLit(v)
	case *ast.Ident:
		return generateCodeFromIdent(v)
	case *ast.SelectorExpr:
		return generateCodeFromSelectorExpr(v)
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
	panic(fmt.Errorf("found unrecognized type: %v", e))
}

func generateCodeFromUnaryExpr(unaryExpr *ast.UnaryExpr) jen.Code {
	return jen.Op(unaryExpr.Op.String()).Add(generateCodeFromExpr(unaryExpr.X))
}

func generateCodeFromBinaryExpr(binaryExpr *ast.BinaryExpr) jen.Code {
	return jen.Add(generateCodeFromExpr(binaryExpr.X)).Op(binaryExpr.Op.String()).Add(generateCodeFromExpr(binaryExpr.Y))
}

func generateCodeFromBasicLit(basicLit *ast.BasicLit) jen.Code {
	switch basicLit.Kind {
	case token.INT:
		i, ok := parseInt(basicLit.Value)
		if !ok {
			panic(fmt.Errorf("invalid_int: %s", basicLit.Value))
		}
		return jen.Lit(i)
	case token.FLOAT:
		panic(fmt.Errorf("not_supported_yet: cannot_convert_float_literal: %v", basicLit))
	case token.CHAR:
		panic(fmt.Errorf("not_supported_yet: cannot_convert_char_literal: %v", basicLit))
	case token.STRING:
		panic(fmt.Errorf("not_supported_yet: cannot_convert_string_literal: %v", basicLit))
	case token.IMAG:
		panic(fmt.Errorf("not_supported_yet: cannot_convert_imag_literal: %v", basicLit))
	}
	panic(fmt.Errorf("cannot_parse_basic_lit: %v", basicLit))
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

func generateCodeFromIdent(ident *ast.Ident) jen.Code {
	switch ident.Name {
	case "bool":
		return jen.Bool()
	case "int":
		return jen.Int()
	case "int8":
		return jen.Int8()
	case "int16":
		return jen.Int16()
	case "int32":
		return jen.Int32()
	case "int64":
		return jen.Int64()
	case "uint":
		return jen.Uint()
	case "uint8":
		return jen.Uint8()
	case "uint16":
		return jen.Uint16()
	case "uint32":
		return jen.Uint32()
	case "uint64":
		return jen.Uint64()
	case "uintptr":
		return jen.Uintptr()
	case "float32":
		return jen.Float32()
	case "float64":
		return jen.Float64()
	case "complex64":
		return jen.Complex64()
	case "complex128":
		return jen.Complex128()
	case "string":
		return jen.String()
	}
	return jen.Id(ident.Name)
}

func generateCodeFromSelectorExpr(selectorExpr *ast.SelectorExpr) jen.Code {
	return jen.Add(generateCodeFromExpr(selectorExpr.X)).Dot(selectorExpr.Sel.Name)
}

func generateCodeFromEllipsis(ellipsis *ast.Ellipsis) jen.Code {
	return jen.Op("...").Add(generateCodeFromExpr(ellipsis.Elt))
}

func generateCodeFromStarExpr(starExpr *ast.StarExpr) jen.Code {
	return jen.Op("*").Add(generateCodeFromExpr(starExpr.X))
}

func generateCodeFromArrayType(arrayType *ast.ArrayType) jen.Code {
	if arrayType.Len == nil {
		return jen.Index().Add(generateCodeFromExpr(arrayType.Elt))
	}
	return jen.Index(generateCodeFromExpr(arrayType.Len)).Add(generateCodeFromExpr(arrayType.Elt))
}

func generateCodeFromFuncType(funcType *ast.FuncType) jen.Code {
	params := generateCodeFromFieldList(funcType.Params)
	results := generateCodeFromFieldList(funcType.Results)
	return jen.Func().Params(params...).Params(results...)
}

func generateCodeFromFieldList(fields *ast.FieldList) []jen.Code {
	params := make([]jen.Code, 0, fields.NumFields())
	for _, field := range fields.List {
		typeCode := generateCodeFromExpr(field.Type)

		if len(field.Names) == 0 {
			params = append(params, typeCode)
			continue
		}

		for _, id := range field.Names {
			params = append(params, jen.Id(id.Name).Add(typeCode))
		}
	}
	return params
}

func generateCodeFromMapType(mapType *ast.MapType) jen.Code {
	keyDef := generateCodeFromExpr(mapType.Key)
	valDef := generateCodeFromExpr(mapType.Value)
	return jen.Map(keyDef).Add(valDef)
}

func generateCodeFromChanType(chanType *ast.ChanType) jen.Code {
	c := generateCodeFromExpr(chanType.Value)
	switch chanType.Dir {
	case ast.RECV:
		return jen.Op("<-").Chan().Add(c)
	case ast.SEND:
		return jen.Chan().Op("<-").Add(c)
	default:
		return jen.Chan().Add(c)
	}
}

func generateCodeFromStructType(structType *ast.StructType) jen.Code {
	params := make([]jen.Code, 0, structType.Fields.NumFields())
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			params = append(
				params,
				jen.Id(name.String()).Add(generateCodeFromExpr(field.Type)),
			)
		}
	}
	return jen.Struct(params...)
}

func generateCodeFromInterfaceType(interfaceType *ast.InterfaceType) jen.Code {
	nMethod := interfaceType.Methods.NumFields()
	methods := make([]jen.Code, 0, nMethod)
	for _, field := range interfaceType.Methods.List {
		name := field.Names[0].String()
		funcType := field.Type.(*ast.FuncType)
		params := generateCodeFromFieldList(funcType.Params)
		results := generateCodeFromFieldList(funcType.Results)
		methods = append(methods, jen.Id(name).Params(params...).Params(results...))
	}
	return jen.Interface(methods...)
}
