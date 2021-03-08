package rx

import (
	"context"
	"reflect"
)

type Observer interface {
	Type() reflect.Type
	OnSubscribe(disposable Disposable)
	OnNext(ctx context.Context, msg interface{})
	OnError(ctx context.Context, err error)
	OnComplete(ctx context.Context)
}

type ObservableOperators interface {
	BlockingForEach(ctx context.Context, consumer interface{}) error
}

type ObservableSource interface {
	Type() reflect.Type
	Subscribe(ctx context.Context, ob Observer)
}

type Observable interface {
	ObservableOperators
	ObservableSource
}

// Emitter acts as a source of signals in push-fashion
type Emitter interface {
	OnNext(ctx context.Context, msg interface{})
	OnError(ctx context.Context, err error)
	OnComplete(ctx context.Context)
}

type ObservableEmitter interface {
	Emitter
	SetDisposable(disposable Disposable)
	IsDisposed() bool
}

type OnSubscribeCall func(ctx context.Context, ob ObservableEmitter)

func Create(onSubscribe OnSubscribeCall) Observable {
	return (&ObservableOnSubscribe{
		onSubscribe: onSubscribe,
	}).Init()
}

func Just(items ...interface{}) Observable {
	return (&ObservableOnSubscribe{
		onSubscribe: func(ctx context.Context, ob ObservableEmitter) {
			for _, item := range items {
				if ob.IsDisposed() {
					return
				}
				ob.OnNext(ctx, item)
			}
			ob.OnComplete(ctx)
		},
	}).Init()
}
