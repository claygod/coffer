package usecases

// Coffer
// Checkpoint
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/claygod/coffer/domain"
)

type checkpoint struct {
	dirPath string
}

func (c *checkpoint) save(repo domain.RecordsRepository, fileName string) error {
	chpName := c.getNewCheckPointName()
	f, err := os.Create(chpName)
	if err != nil {
		return err
	}
	defer f.Close()

	chRecord := make(chan *domain.Record, 10) //TODO: size?
	repo.Iterator(chRecord)                   //     l.store.iterator(chRecord)
	for {
		rec := <-chRecord
		if rec == nil {
			break
		}
		prb, err := c.prepareRecordToCheckpoint(rec.Key, rec.Value)
		if err != nil {
			defer os.Remove(chpName)
			return err
		}
		if _, err := f.Write(prb); err != nil {
			defer os.Remove(chpName)
			return err
		}
	}
	if err := os.Rename(chpName, chpName+"point"); err != nil {
		defer os.Remove(chpName)
		return err
	}
	return nil

}

func (c *checkpoint) load(repo domain.RecordsRepository, fileName string) {

}

func (c *checkpoint) getNewCheckPointName() string {
	for {
		newFileName := c.dirPath + strconv.Itoa(int(time.Now().Unix())) + ".check"
		if _, err := os.Stat(newFileName); !os.IsExist(err) {
			return newFileName
		}
		time.Sleep(1 * time.Second)
	}
}

func (c *checkpoint) prepareRecordToCheckpoint(key string, value []byte) ([]byte, error) {
	if len(key) > maxKeyLength {
		return nil, fmt.Errorf("Key length %d is greater than permissible %d", len(key), maxKeyLength)
	}
	if len(value) > maxValueLength {
		return nil, fmt.Errorf("Value length %d is greater than permissible %d", len(value), maxValueLength)
	}

	var size uint64 = uint64(len([]byte(value)))
	size = size << 16
	size += uint64(len(key))

	return append(uint64ToBytes(size), (append([]byte(key), value...))...), nil
}
