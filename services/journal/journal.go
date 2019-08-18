package journal

// Coffer
// Journal
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/claygod/coffer/services/filenamer"
	"github.com/claygod/tools/batcher"
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
}

func New(cnf *Config, fn *filenamer.FileNamer, alarmFunc func(error)) (*Journal, error) { //TODO: убрать dirPath
	nName, err := fn.GetNewFileName(".log") //dirPath
	if err != nil {
		return nil, err
	}
	clt, err := batcher.Open(nName, cnf.BatchSize)
	if err != nil {
		return nil, err
	}
	return &Journal{
		config:    cnf,
		fileNamer: fn,
		client:    clt,
		//dirPath:   dirPath,
		alarmFunc: alarmFunc,
		//batchSize: batchSize,
	}, nil
}

func (j *Journal) Restart() {
	atomic.StoreInt64(&j.counter, j.config.LimitRecordsPerLogfile+1)
}

func (j *Journal) Write(toSave []byte) {
	clt, err := j.getClient()
	if err != nil {
		j.alarmFunc(err)
	} else {
		clt.Write(toSave)
	}
}

func (j *Journal) Close() {
	j.client.Close()
	for {
		if atomic.LoadInt64(&j.countBatchClients) == 0 {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (j *Journal) getClient() (*batcher.Client, error) {
	j.m.Lock()
	defer j.m.Unlock()
	if j.counter > j.config.LimitRecordsPerLogfile {
		oldClt := j.client
		//fmt.Println("Journal-1", j.fileNamer)
		nName, err := j.fileNamer.GetNewFileName(".log") // j.dirPath
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
		go j.clientBatchClose(oldClt)
	}
	j.counter++
	return j.client, nil
}

func (j *Journal) clientBatchClose(clt *batcher.Client) {
	clt.Close()
	atomic.AddInt64(&j.countBatchClients, -1)
}
