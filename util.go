package gomocker

import (
	"fmt"
	"reflect"

	"github.com/dave/jennifer/jen"
)

func generateDefinitionFromType(t reflect.Type) jen.Code {
	if t.Name() != "" {
		return jen.Qual(t.PkgPath(), t.Name())
	}

	switch t.Kind() {
	case reflect.Ptr:
		return generatePtrDefinitionFromType(t)
	case reflect.Slice:
		return generateSliceDefinitionFromType(t)
	case reflect.Chan:
		return generateChanDefinitionFromType(t)
	case reflect.Struct:
		return generateStructDefinitionFromType(t)
	case reflect.Map:
		return generateMapDefinitionFromType(t)
	case reflect.Array:
		return generateArrayDefinitionFromType(t)
	case reflect.Func:
		return generateFuncDefinitionFromType(t)
	case reflect.Interface:
		return generateInterfaceDefinitionFromType(t)
	case reflect.Bool:
		return jen.Bool()
	case reflect.Int:
		return jen.Int()
	case reflect.Int8:
		return jen.Int8()
	case reflect.Int16:
		return jen.Int16()
	case reflect.Int32:
		return jen.Int32()
	case reflect.Int64:
		return jen.Int64()
	case reflect.Uint:
		return jen.Uint()
	case reflect.Uint8:
		return jen.Uint8()
	case reflect.Uint16:
		return jen.Uint16()
	case reflect.Uint32:
		return jen.Uint32()
	case reflect.Uint64:
		return jen.Uint64()
	case reflect.Uintptr:
		return jen.Uintptr()
	case reflect.Float32:
		return jen.Float32()
	case reflect.Float64:
		return jen.Float64()
	case reflect.Complex64:
		return jen.Complex64()
	case reflect.Complex128:
		return jen.Complex128()
	case reflect.String:
		return jen.String()
	}

	panic(fmt.Errorf("unknown type"))
}

func generatePtrDefinitionFromType(t reflect.Type) jen.Code {
	return jen.Op("*").Add(generateDefinitionFromType(t.Elem()))
}

func generateSliceDefinitionFromType(t reflect.Type) jen.Code {
	return jen.Index().Add(generateDefinitionFromType(t.Elem()))
}

func generateChanDefinitionFromType(t reflect.Type) jen.Code {
	c := generateDefinitionFromType(t.Elem())
	switch t.ChanDir() {
	case reflect.RecvDir:
		return jen.Op("<-").Chan().Add(c)
	case reflect.SendDir:
		return jen.Chan().Op("<-").Add(c)
	default:
		return jen.Chan().Add(c)
	}
}

func generateStructDefinitionFromType(t reflect.Type) jen.Code {
	params := make([]jen.Code, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		params = append(
			params,
			jen.Id(t.Field(i).Name).Add(generateDefinitionFromType(t.Field(i).Type)),
		)
	}
	return jen.Struct(params...)
}

func generateMapDefinitionFromType(t reflect.Type) jen.Code {
	keyDef := generateDefinitionFromType(t.Key())
	valDef := generateDefinitionFromType(t.Elem())
	return jen.Map(keyDef).Add(valDef)
}

func generateArrayDefinitionFromType(t reflect.Type) jen.Code {
	arrayLen := jen.Lit(t.Len())
	elemDef := generateDefinitionFromType(t.Elem())
	return jen.Index(arrayLen).Add(elemDef)
}

func generateFuncDefinitionFromType(t reflect.Type) jen.Code {
	params := make([]jen.Code, t.NumIn(), t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		if i == t.NumIn()-1 && t.IsVariadic() {
			params[i] = jen.Op("...").Add(generateDefinitionFromType(t.In(i).Elem()))
		} else {
			params[i] = generateDefinitionFromType(t.In(i))
		}
	}

	results := make([]jen.Code, 0, t.NumOut())
	for i := 0; i < t.NumOut(); i++ {
		results = append(results, generateDefinitionFromType(t.Out(i)))
	}

	return jen.Func().Params(params...).Params(results...)
}

func generateInterfaceDefinitionFromType(t reflect.Type) jen.Code {
	if t.Name() == "error" && t.PkgPath() == "" {
		return jen.Error()
	}
	if t.Name() != "" {
		return jen.Qual(t.PkgPath(), t.Name())
	}

	nMethod := t.NumMethod()
	methods := make([]jen.Code, nMethod, nMethod)
	for i := 0; i < nMethod; i++ {
		methodType := t.Method(i).Type

		params := make([]jen.Code, 0, methodType.NumIn())
		for i := 0; i < methodType.NumIn(); i++ {
			params = append(params, generateDefinitionFromType(methodType.In(i)))
		}

		results := make([]jen.Code, 0, methodType.NumOut())
		for i := 0; i < methodType.NumOut(); i++ {
			results = append(results, generateDefinitionFromType(methodType.Out(i)))
		}

		methods[i] = jen.Id(t.Method(i).Name).Params(params...).Params(results...)
	}

	return jen.Interface(methods...)
}
