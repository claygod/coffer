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

// func (s *StartStop) Stop() bool {
// 	for i := 0; i < maxIterations; i++ {
// 		if atomic.LoadInt64(&s.enumerator) == stateReady || atomic.CompareAndSwapInt64(&s.enumerator, stateRun, stateReady) {
// 			return true
// 		}
// 		runtime.Gosched()
// 		time.Sleep(s.pause)
// 	}
// 	return false
// }

func (s *StartStop) Stop() bool {
	for i := 0; i < maxIterations; i++ {
		curNum := atomic.LoadInt64(&s.enumerator)
		switch {
		case curNum == -blockedBarrier: // после блокирования все задачи наконец выполнились
			atomic.CompareAndSwapInt64(&s.enumerator, -blockedBarrier, stateReady)
		case curNum < stateBlocked: // не все задачи закончены
		// ждём и надеемся на выполнение всех задач, но новые тут уже точно не появятся
		case curNum == stateBlocked: // заблокировано но и остановлено тоже
			return true
		case curNum == stateReady: // лучший вариант
			return true
		case curNum == stateRun:
			atomic.CompareAndSwapInt64(&s.enumerator, stateRun, stateReady)
		case curNum >= stateRun: // отключаем возможность запуска новых задач
			atomic.CompareAndSwapInt64(&s.enumerator, curNum, curNum-blockedBarrier)
		}
		runtime.Gosched()
		time.Sleep(s.pause)
	}
	return false
}

func (s *StartStop) Block() bool {
	for i := 0; i < maxIterations; i++ {
		if s.Stop() && atomic.CompareAndSwapInt64(&s.enumerator, stateReady, stateBlocked) {
			return true
		}
		runtime.Gosched()
		time.Sleep(s.pause)
	}
	return false
}

func (s *StartStop) Unblock() bool {
	for i := 0; i < maxIterations; i++ {
		if atomic.CompareAndSwapInt64(&s.enumerator, stateBlocked, stateReady) {
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
		if curNum <= stateReady { // заблокировано
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
