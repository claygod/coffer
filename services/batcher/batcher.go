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
type Batcher struct {
	indicator *indicator
	work      io.Writer
	alarm     func(error)
	chInput   chan []byte
	chStop    chan struct{}
	batchSize int
	stopFlag  int64
}

/*
NewBatcher - create new batcher.
Arguments:
	- workFunc	- function that records the formed batch
	- alarmFunc	- error handling function
	- chInput	- input channel
	- batchSize	- batch size
*/
func NewBatcher(workFunc io.Writer, alarmFunc func(error), chInput chan []byte, batchSize int) *Batcher { //TODO: кажется при инициализации батчера не нужно ему давать канал
	return &Batcher{
		indicator: newIndicator(),
		work:      workFunc,
		alarm:     alarmFunc,
		chInput:   make(chan []byte, batchSize),              // chInput,
		chStop:    make(chan struct{}, batchRatio*batchSize), //TODO: тут сдлина может НЕ иметь значение
		batchSize: batchSize,
	}
}

/*
Start - run a worker
*/
func (b *Batcher) Start() {
	if atomic.CompareAndSwapInt64(&b.stopFlag, stateStop, stateStart) {
		go b.indicator.autoSwitcher()
		go b.worker()
	}
}

/*
Stop - finish the job
*/
func (b *Batcher) Stop() {
	close(b.chStop)
	//TODO: может пригодится? b.chStop <- struct{}{}
	for {
		if atomic.LoadInt64(&b.stopFlag) == stateStop {
			return
		}
		runtime.Gosched()
	}
}

/*
GetChan - get current channel.
*/
func (b *Batcher) GetChan() chan struct{} {
	return b.indicator.getChan()
}
