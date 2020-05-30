package gomocker

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

type Type struct {
	PrimitiveType *PrimitiveType
	QualType      *QualType
	ChanType      *ChanType
	SliceType     *SliceType
	PtrType       *PtrType
	ArrayType     *ArrayType
	MapType       *MapType
	FuncType      *FuncType
	StructType    *StructType
	InterfaceType *InterfaceType
}

type PrimitiveKind string

const (
	PrimitiveKindBool       PrimitiveKind = "bool"
	PrimitiveKindByte       PrimitiveKind = "byte"
	PrimitiveKindInt        PrimitiveKind = "int"
	PrimitiveKindInt8       PrimitiveKind = "int8"
	PrimitiveKindInt16      PrimitiveKind = "int16"
	PrimitiveKindInt32      PrimitiveKind = "int32"
	PrimitiveKindInt64      PrimitiveKind = "int64"
	PrimitiveKindUint       PrimitiveKind = "uint"
	PrimitiveKindUint8      PrimitiveKind = "uint8"
	PrimitiveKindUint16     PrimitiveKind = "uint16"
	PrimitiveKindUint32     PrimitiveKind = "uint32"
	PrimitiveKindUint64     PrimitiveKind = "uint64"
	PrimitiveKindUintptr    PrimitiveKind = "uintptr"
	PrimitiveKindFloat32    PrimitiveKind = "float32"
	PrimitiveKindFloat64    PrimitiveKind = "float64"
	PrimitiveKindComplex64  PrimitiveKind = "complex64"
	PrimitiveKindComplex128 PrimitiveKind = "complex128"
	PrimitiveKindString     PrimitiveKind = "string"
	PrimitiveKindError      PrimitiveKind = "error"
)

type PrimitiveType struct {
	Kind PrimitiveKind
}

type QualType struct {
	Package string
	Name    string
}

type ChanTypeDir int

const (
	ChanTypeDirRecv = iota
	ChanTypeDirSend
	ChanTypeDirBoth
)

type ChanType struct {
	Dir  ChanTypeDir
	Elem Type
}

type SliceType struct {
	Elem Type
}

type PtrType struct {
	Elem Type
}

type ArrayType struct {
	Len  int
	Elem Type
}

type MapType struct {
	Key  Type
	Elem Type
}

type TypeField struct {
	Name string
	Type Type
}

type FuncType struct {
	Inputs     []TypeField
	Outputs    []TypeField
	IsVariadic bool
}

type StructType struct {
	Fields []TypeField
}

type InterfaceTypeMethod struct {
	Name string
	Func FuncType
}

type InterfaceType struct {
	Methods []InterfaceTypeMethod
}

func (t PrimitiveType) Type() Type {
	return Type{PrimitiveType: &t}
}

func (t QualType) Type() Type      { return Type{QualType: &t} }
func (t ChanType) Type() Type      { return Type{ChanType: &t} }
func (t SliceType) Type() Type     { return Type{SliceType: &t} }
func (t PtrType) Type() Type       { return Type{PtrType: &t} }
func (t ArrayType) Type() Type     { return Type{ArrayType: &t} }
func (t MapType) Type() Type       { return Type{MapType: &t} }
func (t FuncType) Type() Type      { return Type{FuncType: &t} }
func (t StructType) Type() Type    { return Type{StructType: &t} }
func (t InterfaceType) Type() Type { return Type{InterfaceType: &t} }

type typeCodeGenerator struct{}

var defaultTypeCodeGenerator = &typeCodeGenerator{}

func GenerateCode(typ Type, flag int32) jen.Code {
	return defaultTypeCodeGenerator.GenerateCode(typ, flag)
}

var (
	DefaultFlag      int32 = 0
	BareFunctionFlag int32 = 1
)

func (c *typeCodeGenerator) GenerateCode(typ Type, flag int32) jen.Code {
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

func (c *typeCodeGenerator) generatePrimitiveType(primitiveType PrimitiveType, flag int32) jen.Code {
	switch primitiveType.Kind {
	case PrimitiveKindBool:
		return jen.Bool()
	case PrimitiveKindByte:
		return jen.Byte()
	case PrimitiveKindInt:
		return jen.Int()
	case PrimitiveKindInt8:
		return jen.Int8()
	case PrimitiveKindInt16:
		return jen.Int16()
	case PrimitiveKindInt32:
		return jen.Int32()
	case PrimitiveKindInt64:
		return jen.Int64()
	case PrimitiveKindUint:
		return jen.Uint()
	case PrimitiveKindUint8:
		return jen.Uint8()
	case PrimitiveKindUint16:
		return jen.Uint16()
	case PrimitiveKindUint32:
		return jen.Uint32()
	case PrimitiveKindUint64:
		return jen.Uint64()
	case PrimitiveKindUintptr:
		return jen.Uintptr()
	case PrimitiveKindFloat32:
		return jen.Float32()
	case PrimitiveKindFloat64:
		return jen.Float64()
	case PrimitiveKindComplex64:
		return jen.Complex64()
	case PrimitiveKindComplex128:
		return jen.Complex128()
	case PrimitiveKindString:
		return jen.String()
	case PrimitiveKindError:
		return jen.Error()
	}
	panic(fmt.Errorf("invalid primitive kind: %v", primitiveType.Kind))
}

func (c *typeCodeGenerator) generateQualType(qualType QualType, flag int32) jen.Code {
	return jen.Qual(qualType.Package, qualType.Name)
}

func (c *typeCodeGenerator) generateSliceType(sliceType SliceType, flag int32) jen.Code {
	return jen.Index().Add(c.GenerateCode(sliceType.Elem, flag))
}

func (c *typeCodeGenerator) generateArrayType(arrayType ArrayType, flag int32) jen.Code {
	return jen.Index(jen.Lit(arrayType.Len)).Add(c.GenerateCode(arrayType.Elem, flag))
}

func (c *typeCodeGenerator) generatePtrType(ptrType PtrType, flag int32) jen.Code {
	return jen.Op("*").Add(c.GenerateCode(ptrType.Elem, flag))
}

func (c *typeCodeGenerator) generateChanType(chanType ChanType, flag int32) jen.Code {
	switch chanType.Dir {
	case ChanTypeDirRecv:
		return jen.Op("<-").Chan().Add(c.GenerateCode(chanType.Elem, flag))
	case ChanTypeDirSend:
		return jen.Chan().Op("<-").Add(c.GenerateCode(chanType.Elem, flag))
	case ChanTypeDirBoth:
		return jen.Chan().Add(c.GenerateCode(chanType.Elem, flag))
	}
	panic(fmt.Errorf("invalid type: %v", chanType))
}

func (c *typeCodeGenerator) generateMapType(mapType MapType, flag int32) jen.Code {
	return jen.Map(c.GenerateCode(mapType.Key, flag)).Add(c.GenerateCode(mapType.Elem, flag))
}

func (c *typeCodeGenerator) generateStructType(structType StructType, flag int32) jen.Code {
	fields := make([]jen.Code, 0, len(structType.Fields))
	for _, f := range structType.Fields {
		fields = append(fields, jen.Id(f.Name).Add(c.GenerateCode(f.Type, flag)))
	}

	return jen.Struct(fields...)
}

func (c *typeCodeGenerator) generateFuncType(funcType FuncType, flag int32) jen.Code {
	return jen.Func().Params(
		c.generateFuncInputs(funcType.Inputs, funcType.IsVariadic, flag)...
	).Params(
		c.generateFuncOutputs(funcType.Outputs, flag)...
	)
}

func (c *typeCodeGenerator) generateFuncInputs(inputTypes []TypeField, variadic bool, flag int32) []jen.Code {
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

func (c *typeCodeGenerator) generateFuncOutputs(outputTypes []TypeField, flag int32) []jen.Code {
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

func (c *typeCodeGenerator) generateInterfaceType(interfaceType InterfaceType, flag int32) jen.Code {
	methods := make([]jen.Code, 0, len(interfaceType.Methods))
	for _, m := range interfaceType.Methods {
		methods = append(
			methods,
			jen.Id(m.Name).Params(
				c.generateFuncInputs(m.Func.Inputs, m.Func.IsVariadic, flag)...
			).Params(
				c.generateFuncOutputs(m.Func.Outputs, flag)...,
			),
		)
	}

	return jen.Interface(methods...)
}
