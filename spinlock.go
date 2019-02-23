package main

import (
	"runtime"
	"sync/atomic"
)

type SpinLock struct {
	lock int32
}

func (l *SpinLock) tryLock() bool {
	return atomic.CompareAndSwapInt32(&l.lock, 0, 1)
}

func (l *SpinLock) Unlock() {
	atomic.StoreInt32(&l.lock, 0)
}

func (l *SpinLock) Lock() {
	for !l.tryLock() {
		runtime.Gosched()
	}
}
