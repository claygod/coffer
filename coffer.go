package coffer

// Coffer
// API
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/claygod/coffer/usecases"
)

type Coffer struct {
	logger        usecases.Logger
	dataPath      string
	recInteractor *usecases.RecordsInteractor
	folInteractor *usecases.FollowInteractor
	hasp          int64
}

func New(dataPath string) *Coffer {
	return &Coffer{} //TODO:
}

func (c *Coffer) Start() int64 { // return prev state
	for {
		if atomic.LoadInt64(&c.hasp) == stateStarted {
			return stateStarted
		} else if atomic.CompareAndSwapInt64(&c.hasp, stateStopped, stateStarted) {
			//l.journal = journal.New(l.filePath, mockAlarmHandle, nil, l.batchSize)
			return stateStopped
		}
		runtime.Gosched()
		time.Sleep(1 * time.Millisecond)
	}
}

func (c *Coffer) Stop() int64 { // return prev state
	for {
		if atomic.LoadInt64(&c.hasp) == stateStopped {
			return stateStopped
		} else if atomic.CompareAndSwapInt64(&c.hasp, stateStarted, stateStopped) {
			//l.journal.Close()
			//TODO: остановка фолловера, сохранение
			return stateStarted
		}
		runtime.Gosched()
		time.Sleep(1 * time.Millisecond)
	}
}

/*
SetHandler - add handler. This can be done both before launch and during database operation.
*/
func (c *Coffer) SetHandler(handlerName string, handlerMethod func(interface{}, map[string][]byte) (map[string][]byte, error)) error {
	if atomic.LoadInt64(&c.hasp) == stateStarted {
		return fmt.Errorf("Handles cannot be added while the application is running.")
	}
	//return l.handlers.Set(handlerName, handlerMethod)
	return nil //TODO:
}

// func (c *Coffer) Save() error {
// 	// curState := l.Stop()
// 	// if curState == stateStarted {
// 	// 	defer l.Start()
// 	// }
// 	// chpName := getNewCheckPointName(l.filePath)
// 	// f, err := os.Create(chpName)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// defer f.Close()

// 	// chRecord := make(chan *repo.Record, 10) //TODO: size?
// 	// l.store.iterator(chRecord)
// 	// for {
// 	// 	rec := <-chRecord
// 	// 	if rec == nil {
// 	// 		break
// 	// 	}
// 	// 	prb, err := l.prepareRecordToCheckpoint(rec.Key, rec.Body)
// 	// 	if err != nil {
// 	// 		defer os.Remove(chpName)
// 	// 		return err
// 	// 	}
// 	// 	if _, err := f.Write(prb); err != nil {
// 	// 		defer os.Remove(chpName)
// 	// 		return err
// 	// 	}
// 	// }
// 	// if err := os.Rename(chpName, chpName+"point"); err != nil {
// 	// 	defer os.Remove(chpName)
// 	// 	return err
// 	// }
// 	return nil
// }
