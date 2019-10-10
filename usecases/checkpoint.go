package usecases

// Coffer
// Checkpoint helper
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/claygod/coffer/domain"
)

type checkpoint struct {
	m      sync.Mutex
	config *Config
}

func NewCheckpoint(config *Config) *checkpoint {
	return &checkpoint{
		config: config,
	}
}

func (c *checkpoint) save(repo domain.RecordsRepository, chpName string) error {
	c.m.Lock()
	defer c.m.Unlock()
	f, err := os.Create(chpName)
	if err != nil {
		return err
	}
	err = c.saveToFile(repo, f)
	f.Close()
	if err != nil {
		os.Remove(chpName)
		return err
	}
	if err := os.Rename(chpName, chpName+extPoint); err != nil {
		return fmt.Errorf("%v %v", err, os.Remove(chpName))
	}
	return nil
}

func (c *checkpoint) saveToFile(repo domain.RecordsRepository, f *os.File) error {
	chRecord := make(chan *domain.Record, 10) //TODO: size?
	go repo.Iterator(chRecord)
	for {
		rec := <-chRecord
		if rec == nil {
			break
		}
		prb, err := c.prepareRecordToCheckpoint(rec.Key, rec.Value)
		if err != nil {
			return err
		}
		if _, err := f.Write(prb); err != nil {
			return err
		}
	}
	return nil
}

func (c *checkpoint) load(repo domain.RecordsRepository, fileName string) error {
	f, err := os.Open(fileName) // c.config.DirPath + "/" +
	if err != nil {
		return err
	}
	defer f.Close()
	if err := c.loadFromFile(repo, f); err != nil {
		return err
	}
	return c.loadFromFile(repo, f)
}

func (c *checkpoint) loadFromFile(repo domain.RecordsRepository, f *os.File) error {
	rSize := make([]byte, 8)
	recs := make(map[string][]byte)
	for {
		_, err := f.Read(rSize)
		if err != nil {
			if err == io.EOF {
				break
			}
			repo.Reset()
			return err
		}
		rSuint64 := bytesToUint64(rSize)
		sizeKey := int16(rSuint64)
		sizeValue := rSuint64 >> 16

		key := make([]byte, sizeKey)
		n, err := f.Read(key)
		if err != nil {
			// if err == io.EOF { // тут EOF не должно быть?
			// break
			// }
			return err
		} else if n != int(sizeKey) {
			repo.Reset()
			return fmt.Errorf("The key is not fully loaded (%v)", key)
		}

		value := make([]byte, int(sizeValue))
		n, err = f.Read(value)
		if err != nil {
			// if err == io.EOF { // тут EOF не должно быть?
			// break
			// }
			repo.Reset()
			return err
		} else if n != int(sizeValue) {
			repo.Reset()
			return fmt.Errorf("The value is not fully loaded, (%v)", value)
		}
		recs[string(key)] = value
	}
	repo.WriteListUnsafe(recs)
	return nil
}

func (c *checkpoint) prepareRecordToCheckpoint(key string, value []byte) ([]byte, error) {
	if len(key) > c.config.MaxKeyLength {
		return nil, fmt.Errorf("Key length %d is greater than permissible %d", len(key), c.config.MaxKeyLength)
	}
	if len(value) > c.config.MaxValueLength {
		return nil, fmt.Errorf("Value length %d is greater than permissible %d", len(value), c.config.MaxValueLength)
	}

	var size uint64 = uint64(len([]byte(value)))
	size = size << 16
	size += uint64(len(key))

	return append(uint64ToBytes(size), (append([]byte(key), value...))...), nil
}
