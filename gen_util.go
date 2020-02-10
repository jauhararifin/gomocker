package gomocker

import (
	"github.com/dave/jennifer/jen"
	"reflect"
)

func generateJenFromType(t reflect.Type) jen.Code {
	switch t.Kind() {
	case reflect.Ptr:
		return jen.Op("*").Add(generateJenFromType(t.Elem()))
	case reflect.Slice:
		return jen.Index().Add(generateJenFromType(t.Elem()))
	case reflect.Chan:
		c := generateJenFromType(t.Elem())
		if t.ChanDir() == reflect.RecvDir {
			return jen.Op("<-").Chan().Add(c)
		} else if t.ChanDir() == reflect.SendDir {
			return jen.Chan().Op("<-").Add(c)
		} else {
			return jen.Chan().Add(c)
		}
	case reflect.Struct:
		params := make([]jen.Code, t.NumField(), t.NumField())
		for i := 0; i < t.NumField(); i++ {
			params[i] = jen.Id(t.Field(i).Name).Add(generateJenFromType(t.Field(i).Type))
		}
		return jen.Struct(params...)
	case reflect.Map:
		return jen.Map(generateJenFromType(t.Key())).Add(generateJenFromType(t.Elem()))
	case reflect.Array:
		return jen.Index(jen.Lit(t.Len())).Add(generateJenFromType(t.Elem()))
	case reflect.Func:
		params := make([]jen.Code, t.NumIn(), t.NumIn())
		for i := 0; i < t.NumIn(); i++ {
			params[i] = generateJenFromType(t.In(i))
		}
		results := make([]jen.Code, t.NumOut(), t.NumOut())
		for i := 0; i < t.NumOut(); i++ {
			results[i] = generateJenFromType(t.Out(i))
		}

		return jen.Func().Params(params...).Params(results...)
	case reflect.Interface:
		if len(t.Name()) > 0 {
			break
		}

		methods := make([]jen.Code, t.NumMethod(), t.NumMethod())
		for i := 0; i < t.NumMethod(); i++ {
			methodType := t.Method(i).Type
			params := make([]jen.Code, methodType.NumIn(), methodType.NumIn())
			for i := 0; i < methodType.NumIn(); i++ {
				params[i] = generateJenFromType(methodType.In(i))
			}

			results := make([]jen.Code, methodType.NumOut(), methodType.NumOut())
			for i := 0; i < methodType.NumOut(); i++ {
				results[i] = generateJenFromType(methodType.Out(i))
			}

			methods[i] = jen.Id(t.Method(i).Name).Params(params...).Params(results...)
		}
		return jen.Interface(methods...)
	}
	return jen.Qual(t.PkgPath(), t.Name())
}

func isTypeNullable(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Interface, reflect.Func, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice, reflect.Ptr:
		return true
	default:
		return false
	}
}