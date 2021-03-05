package fun

import (
	"context"
	"fmt"
	"reflect"
)

type Caller interface {
	ReceiveType() reflect.Type
	ReturnType() reflect.Type

	Call(ctx context.Context, in interface{}) (interface{}, error)
}

var _ Caller = (*callerImpl)(nil)

type callerImpl struct {
	runnable
	returnType reflect.Type
	hasError   bool
}

func (s *callerImpl) ReceiveType() reflect.Type {
	return s.receiveType
}

func (s *callerImpl) ReturnType() reflect.Type {
	return s.returnType
}

func (s *callerImpl) Call(ctx context.Context, in interface{}) (interface{}, error) {
	args := s.prepareArguments(ctx, in)
	return s.convertOutput(s.f.Call(args))
}

func (s *callerImpl) convertOutput(values []reflect.Value) (interface{}, error) {
	var ret interface{}
	if !values[0].IsValid() {
		ret = nil
	} else {
		ret = values[0].Interface()
	}
	if s.hasError {
		var err error
		e := values[1]
		if !e.IsValid() || e.IsNil() {
			err = nil
		} else {
			err = e.Interface().(error)
		}
		return ret, err
	} else {
		return ret, nil
	}
}

func CallerOf(call interface{}) (Caller, error) {
	if call == nil {
		return nil, fmt.Errorf("call cannot be nil")
	}
	callValue := reflect.ValueOf(call)
	if callValue.Type().Kind() != reflect.Func {
		return nil, fmt.Errorf("call should bye a function")
	}
	callType := callValue.Type()
	hasContext := false
	var receiveType reflect.Type
	numIn := callType.NumIn()
	switch numIn {
	default:
		return nil, fmt.Errorf("call should have either 1 or 2 arguments but got %d", numIn)
	case 1:
		receiveType = callType.In(0)
	case 2:
		hasContext = true
		firstArgType := callType.In(0)
		if firstArgType != contextType {
			return nil, fmt.Errorf("the first argument should be context.Context but got %s", firstArgType)
		}
		receiveType = callType.In(1)
	}

	hasError := false
	var returnType reflect.Type
	numOut := callType.NumOut()
	switch numOut {
	default:
		return nil, fmt.Errorf("call should return either 1 or 2 values but got %d", numOut)
	case 1:
		returnType = callType.Out(0)
	case 2:
		hasError = true
		returnType = callType.Out(0)
		secondArgType := callType.Out(1)
		if secondArgType != errorType {
			return nil, fmt.Errorf("the second return value can only be error but got %s", secondArgType)
		}
	}
	return &callerImpl{
		runnable: runnable{
			f:           callValue,
			receiveType: receiveType,
			hasContext:  hasContext,
		},
		returnType: returnType,
		hasError:   hasError,
	}, nil
}
