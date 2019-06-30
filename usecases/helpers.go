package usecases

// Coffer
// Checkpoint
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	"encoding/gob"
	"unsafe"
)

func uint64ToBytes(i uint64) []byte {
	x := (*[8]byte)(unsafe.Pointer(&i))
	out := make([]byte, 0, 8)
	out = append(out, x[:]...)
	return out
}

func bytesToUint64(b []byte) uint64 {
	var x [8]byte
	copy(x[:], b[:])
	return *(*uint64)(unsafe.Pointer(&x))
}

// type checkpointName struct {
// 	dirPath string
// }

// func (c *checkpointName) getNewCheckPointName() string {
// 	for {
// 		newFileName := c.dirPath + strconv.Itoa(int(time.Now().Unix())) + ".check"
// 		if _, err := os.Stat(newFileName); !os.IsExist(err) {
// 			return newFileName
// 		}
// 		time.Sleep(1 * time.Second)
// 	}
// }

func bodyOperationEncode(req interface{}, code byte) ([]byte, error) {
	var reqBuf bytes.Buffer
	enc := gob.NewEncoder(&reqBuf)
	if err := enc.Encode(req); err != nil {
		return nil, err
	}
	toSaveLog, err := prepareOperatToLog(code, reqBuf.Bytes())
	if err != nil {
		return nil, err
	}
	return toSaveLog, nil
}

func prepareOperatToLog(code byte, value []byte) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.Write(uint64ToBytes(uint64(len(value) + 1))); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(code); err != nil {
		return nil, err
	}
	if _, err := buf.Write(value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getKeysFromMap(arr map[string][]byte) []string {
	keys := make([]string, 0, len(arr))
	for key, _ := range arr {
		keys = append(keys, key)
	}
	return keys
}
