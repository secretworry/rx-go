package fun

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunnerOf(t *testing.T) {
	t.Run("RunnerOf_should_ReturnTheShapeOfGivenFunction", func(t *testing.T) {
		tests := []struct {
			name   string
			f      interface{}
			err    error
			inType reflect.Type
		}{
			{
				name:   "SimpleFunction",
				f:      func(i int) {},
				inType: reflect.TypeOf((*int)(nil)).Elem(),
			},
			{
				name:   "FunctionWithContext",
				f:      func(ctx context.Context, i int) {},
				inType: reflect.TypeOf((*int)(nil)).Elem(),
			},
			{
				name:   "FunctionWithError",
				f:      func(i int) error { return nil },
				inType: reflect.TypeOf((*int)(nil)).Elem(),
			},
			{
				name:   "FunctionWithContextAndError",
				f:      func(ctx context.Context, i int) error { return nil },
				inType: reflect.TypeOf((*int)(nil)).Elem(),
			},
			{
				name: "EmptyArgument",
				f:    func() {},
				err:  fmt.Errorf("call should have either 1 or 2 arguments but got %d", 0),
			},
			{
				name: "TooManyArguments",
				f:    func(a, b, c int) {},
				err:  fmt.Errorf("call should have either 1 or 2 arguments but got %d", 3),
			},
			{
				name: "InvalidFirstArgument",
				f:    func(a, b int) {},
				err:  fmt.Errorf("the first argument should be context.Context but got int"),
			},
			{
				name: "TooManyReturnValues",
				f: func(a int) (int, int) {
					return a, a
				},
				err: fmt.Errorf("call should return either 0 or 1 values but got 2"),
			},
			{
				name: "InvalidReturnValueType",
				f: func(a int) int {
					return a
				},
				err: fmt.Errorf("the first return value can only be error but got int"),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				sp, err := RunnerOf(tt.f)
				if tt.err != nil {
					assert.EqualError(t, err, tt.err.Error())
					return
				} else if !assert.NoError(t, err) {
					return
				}
				if !assert.NotNil(t, sp, "should not return nil callerImpl") {
					return
				}
				assert.Equal(t, tt.inType, sp.ReceiveType(), "expect receive type %s", tt.inType)
			})
		}
	})
}

func TestRunnerImpl_Run(t *testing.T) {
	t.Run("Run_should_CallAsExpected", func(t *testing.T) {
		tests := []struct {
			name      string
			f         func(called *bool) interface{}
			in        interface{}
			expectErr error
		}{
			{
				name: "SimpleRun",
				f: func(called *bool) interface{} {
					return func(a int) { *called = true }
				},
				in: 3,
			},
			{
				name: "RunWithError",
				f: func(called *bool) interface{} {
					return func(a int) error {
						*called = true
						return errTest
					}
				},
				in:        3,
				expectErr: errTest,
			},
			{
				name: "CallWithContext",
				f: func(called *bool) interface{} {
					return func(ctx context.Context, i int) {
						*called = true
					}
				},
				in: 3,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var called = false
				run := tt.f(&called)
				s, err := RunnerOf(run)
				if !assert.NoError(t, err, "should call CallerOf without error") {
					return
				}
				err = s.Run(context.Background(), tt.in)
				if tt.expectErr != nil {
					assert.EqualError(t, err, tt.expectErr.Error())
					return
				} else if !assert.NoError(t, err, "should Call without error") {
					return
				}
				if !assert.True(t, called, "given function should be called") {
					return
				}
			})
		}
	})
}
