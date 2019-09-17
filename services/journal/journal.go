package journal

// Coffer
// Journal
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/claygod/coffer/services/batcher"
	"github.com/claygod/coffer/services/filenamer"
	//"github.com/claygod/tools/batcher"
)

const limitRecordsPerLogfile int64 = 100000

/*
Journal - transactions logs saver (WAL).
*/
type Journal struct {
	m         sync.Mutex
	config    *Config
	fileNamer *filenamer.FileNamer
	counter   int64
	client    *batcher.Client
	//dirPath           string
	alarmFunc func(error)
	//batchSize         int
	countBatchClients int64
	state             int64
}

func New(cnf *Config, fn *filenamer.FileNamer, alarmFunc func(error)) (*Journal, error) {
	// nName, err := fn.GetNewFileName(".log") //dirPath
	// if err != nil {
	// 	return nil, err
	// }
	// clt, err := batcher.Open(nName, cnf.BatchSize)
	// if err != nil {
	// 	return nil, err
	// }
	return &Journal{
		config:    cnf,
		fileNamer: fn,
		//client:    clt,
		//dirPath:   dirPath,
		alarmFunc: alarmFunc,
		//batchSize: batchSize,
		state: stateStarted,
	}, nil
}

func (j *Journal) Start() error {
	j.m.Lock()
	defer j.m.Unlock()
	nName, err := j.fileNamer.GetNewFileName(".log") //dirPath
	if err != nil {
		return err
	}
	clt, err := batcher.Open(nName, j.config.BatchSize)
	if err != nil {
		return err
	}
	j.client = clt
	return nil
}

func (j *Journal) Stop() {
	j.m.Lock()
	defer j.m.Unlock()
	atomic.CompareAndSwapInt64(&j.state, stateStarted, stateStopped) //если не сработает, значит мы уже стопнуты или в панике
	j.client.Close()
	for {
		if atomic.LoadInt64(&j.countBatchClients) == 0 {
			return
		}
		time.Sleep(1 * time.Millisecond)
	}
}

func (j *Journal) Restart() {
	// j.m.Lock()
	// defer j.m.Unlock()
	//TODO: по идее стейт тут не меняется и его проверять не нужно.
	atomic.StoreInt64(&j.counter, j.config.LimitRecordsPerLogfile+1)
	//j.getClient()
}

func (j *Journal) Write(toSave []byte) error {
	if st := atomic.LoadInt64(&j.state); st != stateStarted {
		return fmt.Errorf("State is `%d` (not started).", st)
	}
	clt, err := j.getClient()
	if err != nil {
		j.alarmFunc(err)
		atomic.StoreInt64(&j.state, statePanic)
		return err
	} else {
		clt.Write(toSave)
	}
	return nil
}

func (j *Journal) getClient() (*batcher.Client, error) {
	j.m.Lock()
	defer j.m.Unlock()
	//fmt.Println("++j *Journal) getClient+++", j.counter, j.config.LimitRecordsPerLogfile)
	if j.counter > j.config.LimitRecordsPerLogfile {
		oldClt := j.client
		nName, err := j.fileNamer.GetNewFileName(".log") // j.dirPath
		//fmt.Println("Journal-1", nName, j.fileNamer)
		if err != nil {
			return nil, err
		}
		clt, err := batcher.Open(nName, j.config.BatchSize)
		if err != nil {
			return nil, err
		}
		j.client = clt
		j.counter = 0
		atomic.AddInt64(&j.countBatchClients, 1)
		j.clientBatchClose(oldClt) //TODO: del GO ?
	}
	j.counter++
	return j.client, nil
}

// func (j *Journal) clientReset() error {
// 	oldClt := j.client
// 	nName, err := j.fileNamer.GetNewFileName(".log") // j.dirPath
// 	if err != nil {
// 		return nil, err
// 	}
// 	clt, err := batcher.Open(nName, j.config.BatchSize)
// 	if err != nil {
// 		return nil, err
// 	}
// 	j.client = clt
// 	j.counter = 0
// 	atomic.AddInt64(&j.countBatchClients, 1)
// 	j.clientBatchClose(oldClt) //TODO: del GO ?
// }

func (j *Journal) clientBatchClose(clt *batcher.Client) {
	clt.Close()
	atomic.AddInt64(&j.countBatchClients, -1)
}
