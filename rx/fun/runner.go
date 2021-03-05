package fun

import (
	"context"
	"fmt"
	"reflect"
)

type Runner interface {
	ReceiveType() reflect.Type
	Run(ctx context.Context, in interface{}) error
}

var _ Runner = (*runnerImpl)(nil)

type runnerImpl struct {
	runnable
	hasError bool
}

func (r *runnerImpl) ReceiveType() reflect.Type {
	return r.receiveType
}

func (r *runnerImpl) Run(ctx context.Context, in interface{}) error {
	args := r.prepareArguments(ctx, in)
	return r.convertOutput(r.f.Call(args))
}

func (r *runnerImpl) convertOutput(values []reflect.Value) error {
	if r.hasError {
		var err error
		v := values[0]
		if !v.IsValid() || v.IsNil() {
			err = nil
		} else {
			err = v.Interface().(error)
		}
		return err
	} else {
		return nil
	}
}
func RunnerOf(run interface{}) (Runner, error) {
	if run == nil {
		return nil, fmt.Errorf("call cannot be nil")
	}
	runValue := reflect.ValueOf(run)
	if runValue.Type().Kind() != reflect.Func {
		return nil, fmt.Errorf("call should bye a function")
	}
	runType := runValue.Type()
	hasContext := false
	var receiveType reflect.Type
	numIn := runType.NumIn()
	switch numIn {
	default:
		return nil, fmt.Errorf("call should have either 1 or 2 arguments but got %d", numIn)
	case 1:
		receiveType = runType.In(0)
	case 2:
		hasContext = true
		firstArgType := runType.In(0)
		if firstArgType != contextType {
			return nil, fmt.Errorf("the first argument should be context.Context but got %s", firstArgType)
		}
		receiveType = runType.In(1)
	}

	hasError := false
	numOut := runType.NumOut()
	switch numOut {
	default:
		return nil, fmt.Errorf("call should return either 0 or 1 values but got %d", numOut)
	case 0:
	case 1:
		hasError = true
		firstRetType := runType.Out(0)
		if firstRetType != errorType {
			return nil, fmt.Errorf("the first return value can only be error but got %s", firstRetType)
		}
	}
	return &runnerImpl{
		runnable: runnable{
			f:           runValue,
			receiveType: receiveType,
			hasContext:  hasContext,
		},
		hasError: hasError,
	}, nil
}
