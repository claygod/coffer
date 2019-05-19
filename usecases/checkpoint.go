package usecases

// Coffer
// Checkpoint
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/claygod/coffer/domain"
)

type checkpoint struct {
	dirPath    string
	logger     Logger
	resControl Resourcer
}

func newCheckpoint(path string, log Logger, rc Resourcer) *checkpoint {
	return &checkpoint{
		dirPath:    path,
		logger:     log,
		resControl: rc,
	}
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

func (c *checkpoint) load(repo domain.RecordsRepository, fileName string) error {
	lastCheckoutName, err := c.getLastCheckPointName()
	if err != nil {
		return err
	}

	fl, err := os.Open(c.dirPath + lastCheckoutName)
	if err != nil {
		return err
	}
	stat, _ := fl.Stat()

	if !c.resControl.GetPermission(stat.Size()) {
		return fmt.Errorf("RecourcesControl: not permission for load operation. File size: %d", stat.Size())
	}
	return c.loadRecordsFromCheckpoint(fl, repo)
}

/*
reset -  so that there are no errors with mismatched time
*/
func (c *checkpoint) reset(repo domain.RecordsRepository, fileName string) error {
	//TODO: del all checkpoints without last
	//TODO: rename last checkpoint to zero-name `0.checkpoint`
	return nil
}

func (c *checkpoint) checkNameLastAndCurrent() error {
	lastCheckoutName, err := c.getLastCheckPointName()
	if err != nil {
		return err
	}

	numStr := strings.Replace(lastCheckoutName, ".checkpoint", "", 0)
	numInt64, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return err
	}

	if curTime := time.Now().Unix(); curTime <= numInt64 {
		return fmt.Errorf("Last name `%s` >= curTime `%d`", lastCheckoutName, curTime)
	}
	return nil
}

func (c *checkpoint) loadRecordsFromCheckpoint(f *os.File, repo domain.RecordsRepository) error {
	rSize := make([]byte, 8)

	//f.Seek(0, 0) //  whence: 0 начало файла, 1 текущее положение, and 2 от конца файла.
	//var m runtime.MemStats
	//runtime.ReadMemStats(&m)

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
			// 	break
			// }
			return err
		} else if n != int(sizeKey) {
			return fmt.Errorf("The key is not fully loaded (%v)", key)
		}

		value := make([]byte, int(sizeValue))
		n, err = f.Read(value)
		if err != nil {
			// if err == io.EOF { // тут EOF не должно быть?
			// 	break
			// }
			return err
		} else if n != int(sizeValue) {
			return fmt.Errorf("The value is not fully loaded, (%v)", value)
		}
		rec := &domain.Record{
			Key:   string(key),
			Value: value,
		}
		repo.SetUnsafeRecord(rec)
		//repo. (string(key), value)
	}
	return nil
}

func (c *checkpoint) getLastCheckPointName() (string, error) {
	chpList, err := c.loadSuffixFilesList(".checkpoint")
	if err != nil || len(chpList) == 0 {
		return "", err
	}
	sort.Strings(chpList)
	return chpList[len(chpList)-1], nil
}

func (c *checkpoint) loadSuffixFilesList(suffix string) ([]string, error) {
	filesList, err := c.loadAllFilesList()
	if err != nil {
		return nil, err
	}

	suffList := make([]string, 0, len(filesList))
	for _, fileName := range filesList {
		if strings.HasSuffix(fileName, suffix) {
			suffList = append(suffList, fileName)
		}
	}
	return suffList, nil
}

func (c *checkpoint) loadAllFilesList() ([]string, error) {
	fl, err := os.Open(c.dirPath)
	if err != nil {
		return nil, err
	}
	defer fl.Close()

	return fl.Readdirnames(-1)
}

func (c *checkpoint) getNewCheckPointName() string {
	for {
		chpNum := time.Now().Unix()
		newFileName := c.dirPath + strconv.Itoa(int(chpNum)) + ".check"
		_, err := os.Stat(newFileName)
		if !os.IsExist(err) {
			return newFileName
		} else if err != nil {
			c.logger.Write(err)
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
