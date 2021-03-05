package fun

import (
	"context"
	"reflect"
)

type runnable struct {
	f           reflect.Value
	receiveType reflect.Type
	hasContext  bool
}

func (r *runnable) prepareArguments(ctx context.Context, in interface{}) []reflect.Value {
	if r.hasContext {
		return []reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(in),
		}
	} else {
		return []reflect.Value{
			reflect.ValueOf(in),
		}
	}
}
