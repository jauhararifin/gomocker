package gomocker

import (
	"fmt"
	"go/types"

	"github.com/dave/jennifer/jen"
)

type typeCodeGenerator struct{}

var defaultTypeCodeGenerator = &typeCodeGenerator{}

func GenerateCode(typ types.Type, flag int32) jen.Code {
	return defaultTypeCodeGenerator.GenerateCode(typ, flag)
}

var (
	DefaultFlag      int32 = 0
	BareFunctionFlag int32 = 1
)

func (c *typeCodeGenerator) GenerateCode(typ types.Type, flag int32) jen.Code {
	switch t := typ.(type) {
	case *types.Basic:
		return c.generatePrimitiveType(t, flag)
	case *types.Named:
		return c.generateQualType(t, flag)
	case *types.Slice:
		return c.generateSliceType(t, flag)
	case *types.Array:
		return c.generateArrayType(t, flag)
	case *types.Pointer:
		return c.generatePtrType(t, flag)
	case *types.Chan:
		return c.generateChanType(t, flag)
	case *types.Map:
		return c.generateMapType(t, flag)
	case *types.Struct:
		return c.generateStructType(t, flag)
	case *types.Signature:
		return c.generateFuncType(t, flag)
	case *types.Interface:
		return c.generateInterfaceType(t, flag)
	}

	panic(fmt.Errorf("invalid type: %v", typ))
}

func (c *typeCodeGenerator) generatePrimitiveType(primitiveType *types.Basic, flag int32) jen.Code {
	switch primitiveType.Kind() {
	case types.Bool:
		return jen.Bool()
	case types.Int:
		return jen.Int()
	case types.Int8:
		return jen.Int8()
	case types.Int16:
		return jen.Int16()
	case types.Int32:
		return jen.Int32()
	case types.Int64:
		return jen.Int64()
	case types.Uint:
		return jen.Uint()
	case types.Uint8:
		return jen.Uint8()
	case types.Uint16:
		return jen.Uint16()
	case types.Uint32:
		return jen.Uint32()
	case types.Uint64:
		return jen.Uint64()
	case types.Uintptr:
		return jen.Uintptr()
	case types.Float32:
		return jen.Float32()
	case types.Float64:
		return jen.Float64()
	case types.Complex64:
		return jen.Complex64()
	case types.Complex128:
		return jen.Complex128()
	case types.String:
		return jen.String()
	}
	panic(fmt.Errorf("invalid primitive kind: %v", primitiveType.Kind()))
}

func (c *typeCodeGenerator) generateQualType(qualType *types.Named, flag int32) jen.Code {
	return jen.Qual(qualType.Obj().Pkg().Path(), qualType.Obj().Name())
}

func (c *typeCodeGenerator) generateSliceType(sliceType *types.Slice, flag int32) jen.Code {
	return jen.Index().Add(c.GenerateCode(sliceType.Elem(), flag))
}

func (c *typeCodeGenerator) generateArrayType(arrayType *types.Array, flag int32) jen.Code {
	return jen.Index(jen.Lit(arrayType.Len())).Add(c.GenerateCode(arrayType.Elem(), flag))
}

func (c *typeCodeGenerator) generatePtrType(ptrType *types.Pointer, flag int32) jen.Code {
	return jen.Op("*").Add(c.GenerateCode(ptrType.Elem(), flag))
}

func (c *typeCodeGenerator) generateChanType(chanType *types.Chan, flag int32) jen.Code {
	switch chanType.Dir() {
	case types.RecvOnly:
		return jen.Op("<-").Chan().Add(c.GenerateCode(chanType.Elem(), flag))
	case types.SendOnly:
		return jen.Chan().Op("<-").Add(c.GenerateCode(chanType.Elem(), flag))
	case types.SendRecv:
		return jen.Chan().Add(c.GenerateCode(chanType.Elem(), flag))
	}
	panic(fmt.Errorf("invalid type: %v", chanType))
}

func (c *typeCodeGenerator) generateMapType(mapType *types.Map, flag int32) jen.Code {
	return jen.Map(c.GenerateCode(mapType.Key(), flag)).Add(c.GenerateCode(mapType.Elem(), flag))
}

func (c *typeCodeGenerator) generateStructType(structType *types.Struct, flag int32) jen.Code {
	fields := make([]jen.Code, 0, structType.NumFields())
	for i := 0; i < structType.NumFields(); i++ {
		f := structType.Field(i)
		fields = append(fields, jen.Id(f.Name()).Add(c.GenerateCode(f.Type(), flag)))
	}

	return jen.Struct(fields...)
}

func (c *typeCodeGenerator) generateFuncType(funcType *types.Signature, flag int32) jen.Code {
	return jen.Func().Params(
		c.generateFuncInputs(funcType.Params(), funcType.Variadic(), flag)...,
	).Params(
		c.generateFuncOutputs(funcType.Results(), flag)...,
	)
}

func (c *typeCodeGenerator) generateFuncInputs(inputTypes *types.Tuple, variadic bool, flag int32) []jen.Code {
	inputs := make([]jen.Code, 0, inputTypes.Len())
	for i := 0; i < inputTypes.Len(); i++ {
		inp := inputTypes.At(i)
		typeCode := c.GenerateCode(inp.Type(), flag)
		if variadic && i == inputTypes.Len()-1 {
			typeCode = jen.Op("...").Add(typeCode)
		}

		if flag&BareFunctionFlag != 0 {
			inputs = append(inputs, typeCode)
		} else {
			inputs = append(inputs, jen.Id(inp.Name()).Add(typeCode))
		}
	}
	return inputs
}

func (c *typeCodeGenerator) generateFuncOutputs(outputTypes *types.Tuple, flag int32) []jen.Code {
	outputs := make([]jen.Code, 0, outputTypes.Len())
	for i := 0; i < outputTypes.Len(); i++ {
		out := outputTypes.At(i)
		if flag*BareFunctionFlag != 0 {
			outputs = append(outputs, c.GenerateCode(out.Type(), flag))
		} else {
			outputs = append(outputs, jen.Id(out.Name()).Add(c.GenerateCode(out.Type(), flag)))
		}
	}
	return outputs
}

func (c *typeCodeGenerator) generateInterfaceType(interfaceType *types.Interface, flag int32) jen.Code {
	return jen.InterfaceFunc(func(g *jen.Group) {
		for i := 0; i < interfaceType.NumMethods(); i++ {
			m := interfaceType.Method(i)
			signature := m.Type().(*types.Signature)
			g.Id(m.Name()).Params(
				c.generateFuncInputs(signature.Params(), signature.Variadic(), flag)...,
			).Params(
				c.generateFuncOutputs(signature.Results(), flag)...,
			)
		}
	})
}
