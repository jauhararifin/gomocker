package gomocker

import (
	"fmt"
	"reflect"
	"testing"
)

type ReflectMocker struct {
	mocker *Mocker
	invocationType,
	paramsType,
	returnsType reflect.Type
	variadic bool
}

func NewReflectMocker(t testing.TB, name string, variadic bool, invocationStruct interface{}) *ReflectMocker {
	mocker := NewMocker(t, name)
	invocationType, paramsType, returnsType := parseInvocationType(variadic, invocationStruct)
	r := &ReflectMocker{
		mocker:         mocker,
		invocationType: invocationType,
		paramsType:     paramsType,
		returnsType:    returnsType,
		variadic:       variadic,
	}
	return r
}

func parseInvocationType(variadic bool, invocationStruct interface{}) (invocationType, paramsType, returnsType reflect.Type) {
	invocationType = reflect.TypeOf(invocationStruct)
	if invocationType.Kind() != reflect.Struct {
		panic(fmt.Errorf("invocation is not a struct"))
	}

	paramsStruct, ok := invocationType.FieldByName("Parameters")
	if !ok {
		panic(fmt.Errorf("missing `Parameters` field in invocation struct"))
	}

	paramsType = paramsStruct.Type
	if paramsType.Kind() != reflect.Struct {
		panic(fmt.Errorf("the invocation's `Parameters` field is not a struct"))
	}

	if variadic && paramsType.Field(paramsType.NumField()-1).Type.Kind() != reflect.Slice {
		panic(fmt.Errorf("the function is variadic but the last field is not a slice"))
	}

	returnsStruct, ok := invocationType.FieldByName("Returns")
	if !ok {
		panic(fmt.Errorf("missing `Returns` field in invocation struct"))
	}

	returnsType = returnsStruct.Type
	if returnsType.Kind() != reflect.Struct {
		panic(fmt.Errorf("the invocation's `Returns` field is not a struct"))
	}

	return
}

func (r *ReflectMocker) Call(params ...interface{}) interface{} {
	returnValues := r.mocker.Call(params...)
	return r.convertReturnsToStruct(returnValues)
}

func (r *ReflectMocker) CallVariadic(params ...interface{}) interface{} {
	actualParams := make([]interface{}, 0, 0 )
	for i := 0; i < len(params)-1; i++ {
		actualParams = append(actualParams, params[i])
	}

	lastVal := reflect.ValueOf(params[len(params)-1])
	if lastVal.Kind() != reflect.Slice {
		panic(fmt.Errorf("last value is not a slice"))
	}
	for i := 0; i < lastVal.Len(); i++ {
		actualParams = append(actualParams, lastVal.Index(i).Interface())
	}

	return r.Call(actualParams...)
}

func (r *ReflectMocker) convertReturnsToStruct(returns []interface{}) interface{} {
	resultPtr := reflect.New(r.returnsType)
	result := resultPtr.Elem()
	for i := 0; i < r.returnsType.NumField(); i++ {
		v := reflect.ValueOf(returns[i])
		if returns[i] != nil {
			result.Field(i).Set(v)
		}
	}
	return result.Interface()
}

func (r *ReflectMocker) MockReturnDefaultValues(nTimes int) {
	returnValues := make([]interface{}, 0, 0)
	for i := 0; i < r.returnsType.NumField(); i++ {
		t := r.returnsType.Field(i).Type
		zeroVal := reflect.New(t).Elem().Interface()
		returnValues = append(returnValues, zeroVal)
	}
	r.mocker.Mock(nTimes, NewFixedReturnsFuncHandler(returnValues...))
}

func (r *ReflectMocker) MockReturnDefaultValuesOnce() {
	r.MockReturnDefaultValues(1)
}

func (r *ReflectMocker) MockReturnDefaultValuesForever() {
	r.MockReturnDefaultValues(LifetimeForever)
}

func (r *ReflectMocker) MockReturnValues(nTimes int, returns ...interface{}) {
	for i := 0; i < r.returnsType.NumField(); i++ {
		if returns[i] == nil && !isTypeNullable(r.returnsType.Field(i).Type) {
			panic("invalid return values type")
		} else if returns[i] != nil {
			v := reflect.ValueOf(returns[i])
			if !v.Type().AssignableTo(r.returnsType.Field(i).Type) {
				panic("invalid return values type")
			}
		}
	}

	r.mocker.Mock(nTimes, NewFixedReturnsFuncHandler(returns...))
}

func (r *ReflectMocker) MockReturnValuesOnce(returns ...interface{}) {
	r.MockReturnValues(1, returns...)
}

func (r *ReflectMocker) MockReturnValuesForever(returns ...interface{}) {
	r.MockReturnValues(LifetimeForever, returns...)
}

func (r *ReflectMocker) Mock(nTimes int, fun interface{}) {
	r.assertFuncSignature(fun)
	funVal := reflect.ValueOf(fun)
	r.mocker.Mock(nTimes, func(parameters ...interface{}) []interface{} {
		inputs := make([]reflect.Value, len(parameters), len(parameters))
		for i, p := range parameters {
			inputs[i] = reflect.ValueOf(p)
		}

		retVals := funVal.Call(inputs)
		outputs := make([]interface{}, len(retVals), len(retVals))
		for i, r := range retVals {
			outputs[i] = r.Interface()
		}

		return outputs
	})
}

func (r *ReflectMocker) assertFuncSignature(fun interface{}) {
	if fun == nil {
		panic(fmt.Errorf("got nil function"))
	}

	funType := reflect.TypeOf(fun)
	if funType.Kind() != reflect.Func {
		panic(fmt.Errorf("fun is not a function"))
	}

	if r.variadic && !funType.IsVariadic() {
		panic(fmt.Errorf("fun is not variadic"))
	}

	if !r.variadic && funType.IsVariadic() {
		panic(fmt.Errorf("fun is variadic"))
	}

	for i := 0; i < funType.NumIn()-1; i++ {
		if !funType.In(i).AssignableTo(r.paramsType.Field(i).Type) {
			panic("wrong function input signature")
		}
	}

	lastField := funType.In(funType.NumIn() - 1)
	if r.variadic && !lastField.AssignableTo(r.paramsType.Field(funType.NumIn()-1).Type) {
		panic("wrong variadic function signature")
	} else if !r.variadic && !lastField.AssignableTo(r.paramsType.Field(funType.NumIn()-1).Type) {
		panic("wrong function input signature")
	}

	for i := 0; i < funType.NumOut(); i++ {
		if !funType.Out(i).AssignableTo(r.returnsType.Field(i).Type) {
			panic("wrong function output signature")
		}
	}
}

func (r *ReflectMocker) MockOnce(fun interface{}) {
	r.Mock(1, fun)
}

func (r *ReflectMocker) MockForever(fun interface{}) {
	r.Mock(LifetimeForever, fun)
}

func (r *ReflectMocker) Invocations() []interface{} {
	invocs := r.mocker.Invocations()

	results := make([]interface{}, 0, 0)
	for _, iv := range invocs {
		results = append(results, r.convertInvocationToStruct(iv))
	}
	return results
}

func (r *ReflectMocker) convertInvocationToStruct(invocation Invocation) interface{} {
	paramsStruct := r.convertParamsToStruct(invocation.Parameters)
	returnsStruct := r.convertReturnsToStruct(invocation.Returns)

	ivcPtr := reflect.New(r.invocationType)
	ivc := ivcPtr.Elem()
	ivc.FieldByName("Parameters").Set(reflect.ValueOf(paramsStruct))
	ivc.FieldByName("Returns").Set(reflect.ValueOf(returnsStruct))

	return ivc.Interface()
}

func (r *ReflectMocker) convertParamsToStruct(params []interface{}) interface{} {
	if r.variadic {
		return r.convertVariadicParamsToStruct(params)
	}

	resultPtr := reflect.New(r.paramsType)
	result := resultPtr.Elem()

	for i := 0; i < r.paramsType.NumField(); i++ {
		v := reflect.ValueOf(params[i])
		if params[i] != nil {
			result.Field(i).Set(v)
		}
	}

	return result.Interface()
}

func (r *ReflectMocker) convertVariadicParamsToStruct(params []interface{}) interface{} {
	resultPtr := reflect.New(r.paramsType)
	result := resultPtr.Elem()

	for i := 0; i < r.paramsType.NumField()-1; i++ {
		v := reflect.ValueOf(params[i])
		if params[i] != nil {
			result.Field(i).Set(v)
		}
	}

	lastFieldIdx := r.paramsType.NumField() - 1
	variadicParams := reflect.MakeSlice(r.paramsType.Field(lastFieldIdx).Type, 0, 0)
	for i := lastFieldIdx; i < len(params); i++ {
		variadicParams = reflect.Append(variadicParams, reflect.ValueOf(params[i]))
	}
	result.Field(lastFieldIdx).Set(variadicParams)

	return result.Interface()
}

func (r *ReflectMocker) TakeOneInvocation() interface{} {
	iv := r.mocker.TakeOneInvocation()
	return r.convertInvocationToStruct(iv)
}
