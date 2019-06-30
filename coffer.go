package coffer

// Coffer
// API
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	//"runtime"
	//"sync/atomic"
	//"time"

	"github.com/claygod/coffer/usecases"
)

type Coffer struct {
	logger        usecases.Logger
	porter        usecases.Porter
	dataPath      string
	recInteractor *usecases.RecordsInteractor
	folInteractor *usecases.FollowInteractor
	hasp          usecases.Starter
}

func New(dataPath string) *Coffer {
	return &Coffer{} //TODO:
}

func (c *Coffer) Start() bool { // return prev state
	return c.hasp.Start()
}

func (c *Coffer) Stop() bool { // return prev state
	return c.hasp.Stop()
}

/*
SetHandler - add handler. This can be done both before launch and during database operation.
*/
func (c *Coffer) SetHandler(handlerName string, handlerMethod func(interface{}, map[string][]byte) (map[string][]byte, error)) error {
	if !c.hasp.IsReady() {
		return fmt.Errorf("Handles cannot be added while the application is running.")
	}

	// if atomic.LoadInt64(&c.hasp) == stateStarted {
	// 	return fmt.Errorf("Handles cannot be added while the application is running.")
	// }
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

func (c *Coffer) checkPanic() {
	if err := recover(); err != nil {
		c.hasp.Block()
		//atomic.StoreInt64(&c.hasp, statePanic)
		fmt.Println(err)
	}
}
