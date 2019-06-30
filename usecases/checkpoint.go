package usecases

// Coffer
// Checkpoint helper
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"io"
	"os"

	"github.com/claygod/coffer/domain"
)

type checkpoint struct {
	dirPath string
}

func (c *checkpoint) save(repo domain.RecordsRepository, chpName string) error {
	//chpName := c.getNewCheckPointName()
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
	if err := os.Rename(chpName, chpName+"point"); err != nil {
		// err2 := os.Remove(chpName)
		return fmt.Errorf("%v %v", err, os.Remove(chpName))
	}
	return nil
}

func (c *checkpoint) saveToFile(repo domain.RecordsRepository, f *os.File) error {
	chRecord := make(chan *domain.Record, 10) //TODO: size?
	repo.Iterator(chRecord)                   // l.store.iterator(chRecord)
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
	f, err := os.Open(c.dirPath + fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.loadFromFile(repo, f)
}

func (c *checkpoint) loadFromFile(repo domain.RecordsRepository, f *os.File) error {
	rSize := make([]byte, 8)
	for {
		_, err := f.Read(rSize)
		if err != nil {
			if err == io.EOF {
				break
			}
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
			return fmt.Errorf("The key is not fully loaded (%v)", key)
		}

		value := make([]byte, int(sizeValue))
		n, err = f.Read(value)
		if err != nil {
			// if err == io.EOF { // тут EOF не должно быть?
			// break
			// }
			return err
		} else if n != int(sizeValue) {
			return fmt.Errorf("The value is not fully loaded, (%v)", value)
		}
		// rec := &domain.Record{
		// 	Key:   string(key),
		// 	Value: value,
		// }
		repo.WriteUnsafeRecord(string(key), value) //          SetUnsafeRecord(rec)
	}
	return nil
}

// func (c *checkpoint) getNewCheckPointName() string {
// 	for {
// 		newFileName := c.dirPath + strconv.Itoa(int(time.Now().Unix())) + ".check"
// 		if _, err := os.Stat(newFileName); !os.IsExist(err) {
// 			return newFileName
// 		}
// 		time.Sleep(1 * time.Second)
// 	}
// }

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

// func bytesToUint64(b []byte) uint64 {
// var x [8]byte
// copy(x[:], b[:])
// return *(*uint64)(unsafe.Pointer(&x))
// }
