package startstop

// Coffer
// StartStop (API)
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"runtime"
	"sync/atomic"
	"time"
)

/*
StartStop - counter to start and stop applications
*/
type StartStop struct {
	enumerator int64
	pause      time.Duration
}

func New(args ...time.Duration) *StartStop {
	pause := pauseDefault
	if len(args) == 1 {
		pause = args[0]
	}
	return &StartStop{
		enumerator: stateReady,
		pause:      pause,
	}
}

func (s *StartStop) Start() bool {
	for i := 0; i < maxIterations; i++ {
		if atomic.LoadInt64(&s.enumerator) == stateRun || atomic.CompareAndSwapInt64(&s.enumerator, stateReady, stateRun) {
			return true
		}
		runtime.Gosched()
		time.Sleep(s.pause)
	}
	return false
}

func (s *StartStop) Stop() bool {
	for i := 0; i < maxIterations; i++ {
		if atomic.LoadInt64(&s.enumerator) == stateReady || atomic.CompareAndSwapInt64(&s.enumerator, stateRun, stateReady) {
			return true
		}
		runtime.Gosched()
		time.Sleep(s.pause)
	}
	return false
}

func (s *StartStop) Add() bool {
	for {
		curNum := atomic.LoadInt64(&s.enumerator)
		if curNum == stateReady {
			return false
		} else if atomic.CompareAndSwapInt64(&s.enumerator, curNum, curNum+1) {
			return true
		}
		runtime.Gosched()
	}
}

func (s *StartStop) Done() bool {
	for {
		curNum := atomic.LoadInt64(&s.enumerator)
		if curNum == stateReady {
			return false
		} else if atomic.CompareAndSwapInt64(&s.enumerator, curNum, curNum-1) {
			return true
		}
		runtime.Gosched()
	}
}

func (s *StartStop) Total() int64 {
	return atomic.LoadInt64(&s.enumerator)
}

func (s *StartStop) IsReady() bool {
	return atomic.LoadInt64(&s.enumerator) == stateReady
}
