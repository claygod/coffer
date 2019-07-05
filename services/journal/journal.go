package journal

// Coffer
// Journal
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	//"io/ioutil"
	//"os"
	//"sort"
	//"strconv"
	//"strings"
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
	m                 sync.Mutex
	fileNamer         *filenamer.FileNamer
	counter           int64
	client            *batcher.Client
	dirPath           string
	alarmFunc         func(error)
	batchSize         int
	countBatchClients int64
}

func New(dirPath string, batchSize int, fn *filenamer.FileNamer, alarmFunc func(error)) (*Journal, error) {
	nName, err := fn.GetNewFileName(dirPath)
	if err != nil {
		return nil, err
	}
	clt, err := batcher.Open(nName, batchSize)
	if err != nil {
		return nil, err
	}
	return &Journal{
		client:    clt,
		dirPath:   dirPath,
		alarmFunc: alarmFunc,
		batchSize: batchSize,
	}, nil
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
		nName, err := j.fileNamer.GetNewFileName(j.dirPath)
		if err != nil {
			return nil, err
		}
		clt, err := batcher.Open(nName, j.batchSize)
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

// func (j *Journal) getNewFileName(dirPath string) (string, error) {
// 	for i := 0; i < 60; i++ {
// 		if latestName, err := j.findLatestLog(); err == nil {

// 		}

// 		newFileName := dirPath + strconv.Itoa(int(time.Now().Unix())) + ".log"
// 		if _, err := os.Stat(newFileName); !os.IsExist(err) {
// 			return newFileName, nil
// 		}
// 		time.Sleep(1 * time.Second)
// 	}
// 	return "", fmt.Errorf("Error finding a new name.")
// }

// func (j *Journal) findLatestLog() (string, error) {
// 	fNamesList, err := j.getFilesByExtList(".log")
// 	if err != nil {
// 		return "", err
// 	}
// 	ln := len(fNamesList)
// 	switch {
// 	case ln == 0:
// 		return "0.log", nil
// 	case ln == 1: // последний лог мы никогда не берём чтобы не ткнуться в ещё наполняемый лог
// 		return fNamesList[0], nil
// 	default:
// 		sort.Strings(fNamesList)
// 		return fNamesList[len(fNamesList)-1], nil
// 	}
// 	//return fNamesList, nil
// }

// func (j *Journal) getFilesByExtList(ext string) ([]string, error) {
// 	files, err := ioutil.ReadDir(j.dirPath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	list := make([]string, 0, len(files))
// 	for _, fl := range files {
// 		if strings.HasSuffix(fl.Name(), ext) {
// 			list = append(list, fl.Name())
// 		}
// 	}
// 	return list, nil
// }
