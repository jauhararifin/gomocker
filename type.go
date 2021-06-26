package gomocker

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/jauhararifin/gotype"
)

type typeCodeGenerator struct{}

var defaultTypeCodeGenerator = &typeCodeGenerator{}

func GenerateCode(typ gotype.Type, flag int32) jen.Code {
	return defaultTypeCodeGenerator.GenerateCode(typ, flag)
}

var (
	DefaultFlag      int32 = 0
	BareFunctionFlag int32 = 1
)

func (c *typeCodeGenerator) GenerateCode(typ gotype.Type, flag int32) jen.Code {
	switch {
	case typ.PrimitiveType != nil:
		return c.generatePrimitiveType(*typ.PrimitiveType, flag)
	case typ.QualType != nil:
		return c.generateQualType(*typ.QualType, flag)
	case typ.SliceType != nil:
		return c.generateSliceType(*typ.SliceType, flag)
	case typ.ArrayType != nil:
		return c.generateArrayType(*typ.ArrayType, flag)
	case typ.PtrType != nil:
		return c.generatePtrType(*typ.PtrType, flag)
	case typ.ChanType != nil:
		return c.generateChanType(*typ.ChanType, flag)
	case typ.MapType != nil:
		return c.generateMapType(*typ.MapType, flag)
	case typ.StructType != nil:
		return c.generateStructType(*typ.StructType, flag)
	case typ.FuncType != nil:
		return c.generateFuncType(*typ.FuncType, flag)
	case typ.InterfaceType != nil:
		return c.generateInterfaceType(*typ.InterfaceType, flag)
	}
	panic(fmt.Errorf("invalid type: %v", typ))
}

func (c *typeCodeGenerator) generatePrimitiveType(primitiveType gotype.PrimitiveType, flag int32) jen.Code {
	switch primitiveType.Kind {
	case gotype.PrimitiveKindBool:
		return jen.Bool()
	case gotype.PrimitiveKindByte:
		return jen.Byte()
	case gotype.PrimitiveKindInt:
		return jen.Int()
	case gotype.PrimitiveKindInt8:
		return jen.Int8()
	case gotype.PrimitiveKindInt16:
		return jen.Int16()
	case gotype.PrimitiveKindInt32:
		return jen.Int32()
	case gotype.PrimitiveKindInt64:
		return jen.Int64()
	case gotype.PrimitiveKindUint:
		return jen.Uint()
	case gotype.PrimitiveKindUint8:
		return jen.Uint8()
	case gotype.PrimitiveKindUint16:
		return jen.Uint16()
	case gotype.PrimitiveKindUint32:
		return jen.Uint32()
	case gotype.PrimitiveKindUint64:
		return jen.Uint64()
	case gotype.PrimitiveKindUintptr:
		return jen.Uintptr()
	case gotype.PrimitiveKindFloat32:
		return jen.Float32()
	case gotype.PrimitiveKindFloat64:
		return jen.Float64()
	case gotype.PrimitiveKindComplex64:
		return jen.Complex64()
	case gotype.PrimitiveKindComplex128:
		return jen.Complex128()
	case gotype.PrimitiveKindString:
		return jen.String()
	case gotype.PrimitiveKindError:
		return jen.Error()
	}
	panic(fmt.Errorf("invalid primitive kind: %v", primitiveType.Kind))
}

func (c *typeCodeGenerator) generateQualType(qualType gotype.QualType, flag int32) jen.Code {
	return jen.Qual(qualType.Package, qualType.Name)
}

func (c *typeCodeGenerator) generateSliceType(sliceType gotype.SliceType, flag int32) jen.Code {
	return jen.Index().Add(c.GenerateCode(sliceType.Elem, flag))
}

func (c *typeCodeGenerator) generateArrayType(arrayType gotype.ArrayType, flag int32) jen.Code {
	return jen.Index(jen.Lit(arrayType.Len)).Add(c.GenerateCode(arrayType.Elem, flag))
}

func (c *typeCodeGenerator) generatePtrType(ptrType gotype.PtrType, flag int32) jen.Code {
	return jen.Op("*").Add(c.GenerateCode(ptrType.Elem, flag))
}

func (c *typeCodeGenerator) generateChanType(chanType gotype.ChanType, flag int32) jen.Code {
	switch chanType.Dir {
	case gotype.ChanTypeDirRecv:
		return jen.Op("<-").Chan().Add(c.GenerateCode(chanType.Elem, flag))
	case gotype.ChanTypeDirSend:
		return jen.Chan().Op("<-").Add(c.GenerateCode(chanType.Elem, flag))
	case gotype.ChanTypeDirBoth:
		return jen.Chan().Add(c.GenerateCode(chanType.Elem, flag))
	}
	panic(fmt.Errorf("invalid type: %v", chanType))
}

func (c *typeCodeGenerator) generateMapType(mapType gotype.MapType, flag int32) jen.Code {
	return jen.Map(c.GenerateCode(mapType.Key, flag)).Add(c.GenerateCode(mapType.Elem, flag))
}

func (c *typeCodeGenerator) generateStructType(structType gotype.StructType, flag int32) jen.Code {
	fields := make([]jen.Code, 0, len(structType.Fields))
	for _, f := range structType.Fields {
		fields = append(fields, jen.Id(f.Name).Add(c.GenerateCode(f.Type, flag)))
	}

	return jen.Struct(fields...)
}

func (c *typeCodeGenerator) generateFuncType(funcType gotype.FuncType, flag int32) jen.Code {
	return jen.Func().Params(
		c.generateFuncInputs(funcType.Inputs, funcType.IsVariadic, flag)...,
	).Params(
		c.generateFuncOutputs(funcType.Outputs, flag)...,
	)
}

func (c *typeCodeGenerator) generateFuncInputs(inputTypes []gotype.TypeField, variadic bool, flag int32) []jen.Code {
	inputs := make([]jen.Code, 0, len(inputTypes))
	for i, inp := range inputTypes {
		typeCode := c.GenerateCode(inp.Type, flag)
		if variadic && i == len(inputTypes)-1 {
			typeCode = jen.Op("...").Add(typeCode)
		}

		if flag&BareFunctionFlag != 0 {
			inputs = append(inputs, typeCode)
		} else {
			inputs = append(inputs, jen.Id(inp.Name).Add(typeCode))
		}
	}
	return inputs
}

func (c *typeCodeGenerator) generateFuncOutputs(outputTypes []gotype.TypeField, flag int32) []jen.Code {
	outputs := make([]jen.Code, 0, len(outputTypes))
	for _, out := range outputTypes {
		if flag*BareFunctionFlag != 0 {
			outputs = append(outputs, c.GenerateCode(out.Type, flag))
		} else {
			outputs = append(outputs, jen.Id(out.Name).Add(c.GenerateCode(out.Type, flag)))
		}
	}
	return outputs
}

func (c *typeCodeGenerator) generateInterfaceType(interfaceType gotype.InterfaceType, flag int32) jen.Code {
	return jen.InterfaceFunc(func(g *jen.Group) {
		for _, m := range interfaceType.Methods {
			g.Id(m.Name).Params(
				c.generateFuncInputs(m.Func.Inputs, m.Func.IsVariadic, flag)...,
			).Params(
				c.generateFuncOutputs(m.Func.Outputs, flag)...,
			)
		}
	})
}
