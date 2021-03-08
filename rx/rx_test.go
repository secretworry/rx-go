package rx

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func SliceConsumer(target interface{}) func(ctx context.Context, ele interface{}) error {
	if target == nil {
		panic("target should not be nil")
	}
	v := reflect.ValueOf(target)
	t := v.Type()
	if t.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("pointer to a slice is required but got %T", target))
	}
	if t.Elem().Kind() != reflect.Slice {
		panic(fmt.Sprintf("pointer to a slice is required but got %T", target))
	}
	return func(ctx context.Context, ele interface{}) error {
		v.Elem().Set(reflect.Append(v.Elem(), reflect.ValueOf(ele)))
		return nil
	}
}

func TestCreate(t *testing.T) {
	t.Run("Create_should_CreateAObservableWithoutError", func(t *testing.T) {
		_ = Create(func(ctx context.Context, ob ObservableEmitter) {
			ob.OnComplete(ctx)
		})
	})
	t.Run("Create_should_EmitElementsAsExpected", func(t *testing.T) {
		expect := []int{1, 2, 3}
		actual := new([]int)
		err := Create(func(ctx context.Context, ob ObservableEmitter) {
			for _, i := range expect {
				if ob.IsDisposed() {
					break
				}
				ob.OnNext(ctx, i)
			}
			ob.OnComplete(ctx)
		}).BlockingForEach(context.Background(), SliceConsumer(actual))
		if !assert.NoError(t, err, "should BlockingForEachObserver without error") {
			return
		}
		if !assert.Equal(t, expect, *actual) {
			return
		}
	})
}

func TestJust(t *testing.T) {
	t.Run("Just_should_CreateObservableWithoutError", func(t *testing.T) {
		_ = Just()
	})

	t.Run("Just_should_CreateObservableEmittingGivenItems", func(t *testing.T) {
		ctx := context.Background()
		expect := []int{1, 2, 3}
		ret := new([]int)
		err := Just(1, 2, 3).BlockingForEach(ctx, SliceConsumer(ret))
		if !assert.NoError(t, err, "should BlockingForEachObserver without error") {
			return
		}
		if !assert.Equal(t, expect, *ret, "should emit given items") {
			return
		}
	})
}
