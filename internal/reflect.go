package internal

import (
	"iter"
	"reflect"
)

// prepareArgs prepares arguments for function invocation via reflection.
// This is an internal helper that shouldn't be exposed to users.
func prepareArgs[T any](fnType reflect.Type, val T) []reflect.Value {
	numIn := fnType.NumIn()

	if numIn == 0 {
		return []reflect.Value{}
	}

	if numIn == 1 {
		firstParamType := fnType.In(0)

		if firstParamType.Kind() == reflect.Interface {
			return []reflect.Value{reflect.ValueOf(val)}
		}

		if fnType.IsVariadic() {
			return []reflect.Value{reflect.ValueOf(val)}
		}

		valType := reflect.TypeOf(val)
		if valType.AssignableTo(firstParamType) {
			return []reflect.Value{reflect.ValueOf(val)}
		}
	}

	return []reflect.Value{reflect.ValueOf(val)}
}

// ExecuteForEach executes a function for each element using reflection.
// Returns an error if the provided argument is not a function.
// This function is exported for use by the parent flow package,
// but cannot be imported by external packages due to internal/ protection.
func ExecuteForEach[T any](source iter.Seq[T], fn any) error {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func {
		return &invalidFunctionError{Fn: fn}
	}

	for val := range source {
		args := prepareArgs(fnType, val)
		fnValue.Call(args)
	}

	return nil
}

// invalidFunctionError is returned when ForEach receives a non-function argument.
type invalidFunctionError struct {
	Fn any
}

func (e *invalidFunctionError) Error() string {
	return "ForEach: argument must be a function"
}
