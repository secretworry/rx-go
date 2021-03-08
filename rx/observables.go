package rx

import (
	"context"
	"reflect"
	"unsafe"

	"www.github.com/secretworry/rx-go/rx/fun"
)

var _ ObservableOperators = (*BaseObservable)(nil)

type BaseObservable struct {
	Self func() ObservableSource
}

func (b BaseObservable) Type() reflect.Type {
	return b.Self().Type()
}

func (b BaseObservable) BlockingForEach(ctx context.Context, consumer interface{}) error {
	runner, err := fun.RunnerOf(consumer)
	if err != nil {
		return err
	}
	ob := NewBlockingForEachObserver(runner)
	source := b.Self()
	source.Subscribe(ctx, ob)
	return ob.Wait(ctx)
}

var _ Observable = (*ObservableOnSubscribe)(nil)

type ObservableOnSubscribe struct {
	BaseObservable
	onSubscribe OnSubscribeCall
}

func (o *ObservableOnSubscribe) Init() *ObservableOnSubscribe {
	o.Self = func() ObservableSource {
		return o
	}
	return o
}

func (o *ObservableOnSubscribe) Subscribe(ctx context.Context, ob Observer) {
	emitter := &createEmitter{ob: ob}
	ob.OnSubscribe(emitter)
	defer func() {
		if e := recover(); e != nil {
			emitter.OnError(ctx, toError(e))
		}
	}()
	o.onSubscribe(ctx, emitter)
}

var _ Disposable = (*createEmitter)(nil)
var _ ObservableEmitter = (*createEmitter)(nil)

type createEmitter struct {
	disposable unsafe.Pointer
	ob         Observer
}

func (e *createEmitter) SetDisposable(disposable Disposable) {
	DisposableHelper.Set(&e.disposable, &disposable)
}

func (e *createEmitter) Dispose() {
	DisposableHelper.Dispose(&e.disposable)
}

func (e *createEmitter) IsDisposed() bool {
	return DisposableHelper.IsDisposed(&e.disposable)
}

func (e *createEmitter) OnNext(ctx context.Context, msg interface{}) {
	if !e.IsDisposed() && !isDone(ctx) {
		e.ob.OnNext(ctx, msg)
	}
}

func (e *createEmitter) OnError(ctx context.Context, err error) {
	if !e.IsDisposed() && !isDone(ctx) {
		e.ob.OnError(ctx, err)
		e.Dispose()
	}
}

func (e *createEmitter) OnComplete(ctx context.Context) {
	if !e.IsDisposed() && !isDone(ctx) {
		e.ob.OnComplete(ctx)
		e.Dispose()
	}
}
