package journal

// Coffer
// Journal
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/claygod/tools/batcher"
)

const limitRecordsPerLogfile int64 = 100000

/*
Journal - transactions logs saver (WAL).
*/
type Journal struct {
	m                 sync.Mutex
	counter           int64
	client            *batcher.Client
	dirPath           string
	alarmFunc         func(error)
	batchSize         int
	countBatchClients int64
}

func New(dirPath string, alarmFunc func(error), chInput chan []byte, batchSize int) *Journal {
	clt, _ := batcher.Open(getNewFileName(dirPath), batchSize)
	return &Journal{
		client:    clt,
		dirPath:   dirPath,
		alarmFunc: alarmFunc,
		batchSize: batchSize,
	}
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
	if j.counter > limitRecordsPerLogfile {
		oldClt := j.client
		clt, err := batcher.Open(getNewFileName(j.dirPath), j.batchSize)
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

func getNewFileName(dirPath string) string {
	for {
		newFileName := dirPath + strconv.Itoa(int(time.Now().Unix())) + ".log"
		if _, err := os.Stat(newFileName); !os.IsExist(err) {
			return newFileName
		}
		time.Sleep(1 * time.Second)
	}
}
