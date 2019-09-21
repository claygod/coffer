package batcher

// Batcher
// API
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"io"
	"runtime"
	"sync/atomic"
)

const batchRatio int = 10 // how many batches fit into the input channel

const (
	stateStop int64 = 0 << iota
	stateStart
)

/*
Batcher - performs write jobs in batches.
*/
type batcher struct {
	indicator *indicator
	work      io.Writer
	alarm     func(error)
	chInput   chan []byte
	chStop    chan struct{}
	batchSize int
	stopFlag  int64
}

/*
newBatcher - create new batcher.
Arguments:
	- workFunc	- function that records the formed batch
	- alarmFunc	- error handling function
	- chInput	- input channel
	- batchSize	- batch size
*/
func newBatcher(workFunc io.Writer, alarmFunc func(error), chInput chan []byte, batchSize int) *batcher { //TODO: кажется при инициализации батчера не нужно ему давать канал
	return &batcher{
		indicator: newIndicator(),
		work:      workFunc,
		alarm:     alarmFunc,
		chInput:   make(chan []byte, batchSize),              // chInput,
		chStop:    make(chan struct{}, batchRatio*batchSize), //TODO: тут длина может НЕ иметь значение
		batchSize: batchSize,
	}
}

/*
start - run a worker
*/
func (b *batcher) start() {
	if atomic.CompareAndSwapInt64(&b.stopFlag, stateStop, stateStart) {
		go b.indicator.autoSwitcher()
		go b.worker()
	}
}

/*
stop - finish the job
*/
func (b *batcher) stop() {
	// if _, ok := <-b.chStop; ok {
	// 	close(b.chStop)
	// }
	if b.chStop != nil {
		close(b.chStop)
	}
	//TODO: может пригодится? b.chStop <- struct{}{}
	for {
		if atomic.LoadInt64(&b.stopFlag) == stateStop {
			return
		}
		runtime.Gosched()
	}
}

/*
getChan - get current channel.
*/
func (b *batcher) getChan() chan struct{} {
	return b.indicator.getChan()
}
