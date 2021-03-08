package rx

import (
	"context"
	"reflect"
	"unsafe"

	"www.github.com/secretworry/rx-go/rx/fun"
)

var _ Disposable = (*BlockingForEachObserver)(nil)
var _ Observer = (*BlockingForEachObserver)(nil)

type BlockingForEachObserver struct {
	disposable unsafe.Pointer
	consumer   fun.Runner
	notify     chan error
}

func NewBlockingForEachObserver(consumer fun.Runner) *BlockingForEachObserver {
	return &BlockingForEachObserver{
		consumer: consumer,
		notify:   make(chan error, 1),
	}
}

func (f BlockingForEachObserver) Type() reflect.Type {
	return f.consumer.ReceiveType()
}

func (f *BlockingForEachObserver) Dispose() {
	DisposableHelper.Dispose(&f.disposable)
}

func (f *BlockingForEachObserver) IsDisposed() bool {
	return DisposableHelper.IsDisposed(&f.disposable)
}

func (f *BlockingForEachObserver) OnSubscribe(disposable Disposable) {
	DisposableHelper.SetOnce(&f.disposable, &disposable)
}

func (f *BlockingForEachObserver) OnNext(ctx context.Context, msg interface{}) {
	if !f.IsDisposed() && !isDone(ctx) {
		err := f.consumer.Run(ctx, msg)
		if err != nil {
			f.dispose(err)
		}
	}
}

func (f *BlockingForEachObserver) OnError(ctx context.Context, err error) {
	if !f.IsDisposed() && !isDone(ctx) {
		f.dispose(err)
	}
}

func (f *BlockingForEachObserver) dispose(err error) {
	if DisposableHelper.Dispose(&f.disposable) {
		f.notify <- err
		close(f.notify)
	}
}

func (f *BlockingForEachObserver) OnComplete(ctx context.Context) {
	if !f.IsDisposed() && !isDone(ctx) {
		f.dispose(nil)
	}
}

func (f *BlockingForEachObserver) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case err := <-f.notify:
		return err
	}
}
