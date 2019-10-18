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

//const limitRecordsPerLogfile int64 = 100000

/*
Journal - transactions logs saver (WAL).
*/
type Journal struct {
	m                 sync.Mutex
	config            *Config
	fileNamer         *filenamer.FileNamer
	counter           int64
	client            *batcher.Client
	alarmFunc         func(error)
	countBatchClients int64
	state             int64
}

/*
New - create new Journal.
*/
func New(cnf *Config, fn *filenamer.FileNamer, alarmFunc func(error)) (*Journal, error) {
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

/*
Start - launch the journal.
*/
func (j *Journal) Start() error {
	j.m.Lock()
	defer j.m.Unlock()
	nName, err := j.fileNamer.GetNewFileName(".log") //dirPath
	if err != nil {
		return err
	}
	clt, err := batcher.Open(nName, j.config.BatchSize, j.alarmFunc)
	if err != nil {
		return err
	}
	j.client = clt
	return nil
}

/*
Stop - stop the journal.
*/
func (j *Journal) Stop() {
	j.m.Lock()
	defer j.m.Unlock()
	atomic.CompareAndSwapInt64(&j.state, stateStarted, stateStopped) // if it doesn’t work, then we are already stopped or in a panic
	j.client.Close()
	for {
		if atomic.LoadInt64(&j.countBatchClients) == 0 {
			return
		}
		time.Sleep(1 * time.Millisecond)
	}
}

/*
Restart - restart the Journal. The counter is set so that the next write is in a new file.
*/
func (j *Journal) Restart() {
	// j.m.Lock()
	// defer j.m.Unlock()
	//TODO: in theory, the state does not change here and you do not need to check it.
	atomic.StoreInt64(&j.counter, j.config.LimitRecordsPerLogfile+1)
}

/*
Write - write data to log file.
*/
func (j *Journal) Write(toSave []byte) error {
	if st := atomic.LoadInt64(&j.state); st != stateStarted {
		return fmt.Errorf("State is `%d` (not started).", st)
	}
	clt, err := j.getClient()
	if err != nil {
		j.alarmFunc(err)
		atomic.StoreInt64(&j.state, statePanic)
		return err
	}
	clt.Write(toSave)

	return nil
}

func (j *Journal) getClient() (*batcher.Client, error) {
	j.m.Lock()
	defer j.m.Unlock()
	if j.counter > j.config.LimitRecordsPerLogfile {
		oldClt := j.client
		nName, err := j.fileNamer.GetNewFileName(".log") // j.dirPath
		if err != nil {
			return nil, err
		}
		clt, err := batcher.Open(nName, j.config.BatchSize, j.alarmFunc)
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

func (j *Journal) clientBatchClose(clt *batcher.Client) {
	clt.Close()
	atomic.AddInt64(&j.countBatchClients, -1)
}
