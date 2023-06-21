package engine

import (
	"fmt"
	"reflect"
)

func InvokeOn(any reflect.Value, name string, args ...interface{}) (Results []reflect.Value, err error) {
	method := any.MethodByName(name)
	methodType := method.Type()
	numIn := methodType.NumIn()
	if numIn > len(args) {
		return nil, fmt.Errorf("Method %s must have minimum %d params. Have %d", name, numIn, len(args))
	}
	if numIn != len(args) && !methodType.IsVariadic() {
		return nil, fmt.Errorf("Method %s must have %d params. Have %d", name, numIn, len(args))
	}
	in := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		var inType reflect.Type
		if methodType.IsVariadic() && i >= numIn-1 {
			inType = methodType.In(numIn - 1).Elem()
		} else {
			inType = methodType.In(i)
		}
		argValue := reflect.ValueOf(args[i])
		if args[i] == nil {
			in[i] = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())
			continue
		}

		if !argValue.IsValid() {
			return nil, fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", name, i, inType, argValue.String())
		}
		argType := argValue.Type()
		if argType.ConvertibleTo(inType) {
			in[i] = argValue.Convert(inType)
		} else {
			return nil, fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", name, i, inType, argType)
		}
	}
	return method.Call(in), nil
}

func Invoke(any interface{}, name string, args ...interface{}) (Results []reflect.Value, err error) {
	v := reflect.ValueOf(any)
	return InvokeOn(v, name, args...)
}

func Invoke1Res1ErrOn(any reflect.Value, name string, args ...interface{}) (Result reflect.Value, err error) {
	results, err := InvokeOn(any, name, args...)
	if err != nil {
		return reflect.ValueOf(nil), err
	}
	if len(results) != 2 {
		return reflect.ValueOf(nil), fmt.Errorf("Method return %d vars instead of 2", len(results))
	}
	if results[1].IsNil() {
		return results[0], nil
	}
	return results[0], results[1].Interface().(error)
}

func Invoke1Res1Err(any interface{}, name string, args ...interface{}) (Result reflect.Value, err error) {
	results, err := Invoke(any, name, args...)
	if err != nil {
		return reflect.ValueOf(nil), err
	}
	if len(results) != 2 {
		return reflect.ValueOf(nil), fmt.Errorf("Method return %d vars instead of 2", len(results))
	}
	if results[1].IsNil() {
		return results[0], nil
	}
	return results[0], results[1].Interface().(error)
}

func HasMethodOn(any reflect.Value, name string) bool {
	m := any.MethodByName(name)
	return m.IsValid() && !m.IsNil()
}

func Invoke1ErrOn(any reflect.Value, name string, args ...interface{}) error {
	results, err := InvokeOn(any, name, args...)
	if err != nil {
		return err
	}
	if len(results) != 1 {
		return fmt.Errorf("Method return %d vars instead of 1", len(results))
	}
	if results[0].IsNil() {
		return nil
	}
	return results[0].Interface().(error)
}
