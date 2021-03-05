package fun

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errTest = fmt.Errorf("test")

func TestCallerOf(t *testing.T) {
	t.Run("CallerOf_should_ReturnTheShapeOfGivenFunction", func(t *testing.T) {
		tests := []struct {
			name    string
			f       interface{}
			err     error
			inType  reflect.Type
			outType reflect.Type
		}{
			{
				name: "SimpleFunction",
				f: func(i int) int {
					return i
				},
				inType:  reflect.TypeOf((*int)(nil)).Elem(),
				outType: reflect.TypeOf((*int)(nil)).Elem(),
			},
			{
				name: "FunctionWithContext",
				f: func(ctx context.Context, i int) int {
					return i
				},
				inType:  reflect.TypeOf((*int)(nil)).Elem(),
				outType: reflect.TypeOf((*int)(nil)).Elem(),
			},
			{
				name: "FunctionWithError",
				f: func(i int) (int, error) {
					return i, nil
				},
				inType:  reflect.TypeOf((*int)(nil)).Elem(),
				outType: reflect.TypeOf((*int)(nil)).Elem(),
			},
			{
				name: "FunctionWithContextAndError",
				f: func(ctx context.Context, i int) (int, error) {
					return i, nil
				},
				inType:  reflect.TypeOf((*int)(nil)).Elem(),
				outType: reflect.TypeOf((*int)(nil)).Elem(),
			},
			{
				name: "EmptyArgument",
				f: func() int {
					return 1
				},
				err: fmt.Errorf("call should have either 1 or 2 arguments but got %d", 0),
			},
			{
				name: "TooManyArguments",
				f: func(a, b, c int) int {
					return a
				},
				err: fmt.Errorf("call should have either 1 or 2 arguments but got %d", 3),
			},
			{
				name: "InvalidFirstArgument",
				f: func(a, b int) int {
					return a
				},
				err: fmt.Errorf("the first argument should be context.Context but got int"),
			},
			{
				name: "EmptyReturnValue",
				f: func(int) {
				},
				err: fmt.Errorf("call should return either 1 or 2 values but got 0"),
			},
			{
				name: "TooManyReturnValues",
				f: func(a int) (int, int, int) {
					return a, a, a
				},
				err: fmt.Errorf("call should return either 1 or 2 values but got 3"),
			},
			{
				name: "InvalidSecondReturnValue",
				f: func(a int) (int, int) {
					return a, a
				},
				err: fmt.Errorf("the second return value can only be error but got int"),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				sp, err := CallerOf(tt.f)
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
				assert.Equal(t, tt.outType, sp.ReturnType(), "expect return type %s", tt.outType)
			})
		}
	})
}

func TestCallerImpl_Call(t *testing.T) {
	t.Run("Call_should_CallAsExpected", func(t *testing.T) {
		tests := []struct {
			name      string
			f         interface{}
			in        interface{}
			expect    interface{}
			expectErr error
		}{
			{
				name:   "SimpleCall",
				f:      func(a int) int { return a },
				in:     3,
				expect: 3,
			},
			{
				name:      "CallWithError",
				f:         func(a int) (int, error) { return 0, errTest },
				in:        3,
				expectErr: errTest,
			},
			{
				name: "CallWithContext",
				f: func(ctx context.Context, i int) int {
					return i
				},
				in:     3,
				expect: 3,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				s, err := CallerOf(tt.f)
				if !assert.NoError(t, err, "should call CallerOf without error") {
					return
				}
				value, err := s.Call(context.Background(), tt.in)
				if tt.expectErr != nil {
					assert.EqualError(t, err, tt.expectErr.Error())
					return
				} else if !assert.NoError(t, err, "should Call without error") {
					return
				}
				if !assert.Equal(t, tt.expect, value) {
					return
				}
			})
		}
	})
}
