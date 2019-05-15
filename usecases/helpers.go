package usecases

// Coffer
// Checkpoint
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
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
