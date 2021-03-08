package rx

import (
	"io"
	"sync/atomic"
	"unsafe"
)

type Disposable interface {
	Dispose()
	IsDisposed() bool
}

var Disposables = struct {
	Disposed   func() Disposable
	Empty      func() Disposable
	FromFunc   func(onDisposed func()) Disposable
	FromCloser func(closer io.Closer) Disposable
}{
	Disposed: func() Disposable {
		return disposedDisposableInstance
	},
	Empty: func() Disposable {
		return &atomicDisposable{}
	},
	FromFunc: func(onDisposed func()) Disposable {
		return &atomicDisposable{
			onDisposed: onDisposed,
		}
	},
	FromCloser: func(closer io.Closer) Disposable {
		return &atomicDisposable{
			onDisposed: func() {
				_ = closer.Close() // we ignore the error intentionally
			},
		}
	},
}

var disposedDisposableInstance = disposedDisposable{}

var _ Disposable = (*disposedDisposable)(nil)

type disposedDisposable struct {
}

func (d disposedDisposable) Dispose() {
}

func (d disposedDisposable) IsDisposed() bool {
	return true
}

var _ Disposable = (*atomicDisposable)(nil)

type atomicDisposable struct {
	disposed   int32
	onDisposed func()
}

func (a *atomicDisposable) Dispose() {
	if atomic.CompareAndSwapInt32(&a.disposed, 0, 1) {
		if a.onDisposed != nil {
			a.onDisposed()
		}
	}
}

func (a *atomicDisposable) IsDisposed() bool {
	return atomic.LoadInt32(&a.disposed) > 0
}

var DISPOSED = &[]Disposable{disposedDisposable{}}[0]

func disposableHelperIsDisposed(ptr *unsafe.Pointer) bool {
	curPtr := atomic.LoadPointer(ptr)
	cur := (*Disposable)(curPtr)
	return cur == DISPOSED
}

var DisposableHelper = struct {
	IsDisposed func(ptr *unsafe.Pointer) bool
	Set        func(ptr *unsafe.Pointer, d *Disposable) bool
	SetOnce    func(ptr *unsafe.Pointer, d *Disposable) bool
	Dispose    func(ptr *unsafe.Pointer) bool
	Replace    func(ptr *unsafe.Pointer, d *Disposable) bool
}{
	IsDisposed: disposableHelperIsDisposed,
	Set: func(ptr *unsafe.Pointer, d *Disposable) bool {
		for {
			curPtr := atomic.LoadPointer(ptr)
			cur := (*Disposable)(curPtr)
			if cur == DISPOSED {
				if d != nil {
					(*d).Dispose()
				}
				return false
			}
			if atomic.CompareAndSwapPointer(ptr, curPtr, unsafe.Pointer(d)) {
				if cur != nil {
					(*cur).Dispose()
				}
				return true
			}
		}
	},
	SetOnce: func(ptr *unsafe.Pointer, d *Disposable) bool {
		if !atomic.CompareAndSwapPointer(ptr, unsafe.Pointer(nil), unsafe.Pointer(d)) {
			(*d).Dispose()
			return false
		}
		return true
	},
	Dispose: func(ptr *unsafe.Pointer) bool {
		curPtr := atomic.LoadPointer(ptr)
		cur := (*Disposable)(curPtr)
		d := DISPOSED
		if cur != d {
			curPtr = atomic.SwapPointer(ptr, unsafe.Pointer(d))
			cur = (*Disposable)(curPtr)
			if cur != d {
				if cur != nil {
					(*cur).Dispose()
				}
				return true
			}
		}
		return false
	},
	Replace: func(ptr *unsafe.Pointer, d *Disposable) bool {
		for {
			curPtr := atomic.LoadPointer(ptr)
			cur := (*Disposable)(curPtr)
			if cur == DISPOSED {
				if d != nil {
					(*d).Dispose()
				}
				return false
			}
			if atomic.CompareAndSwapPointer(ptr, curPtr, unsafe.Pointer(d)) {
				return true
			}
		}
	},
}
