package coffer

// Coffer
// Helpers
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
)

func (c *Coffer) copySlice(inList []string) ([]string, error) { // на случай, если мы хотим скопировать входные данные запроса, боясь их изменения
	outList := make([]string, len(inList))
	n := copy(outList, inList)
	if n != len(inList) {
		return nil, fmt.Errorf("Slice (strings) copy failed.")
	}
	return outList, nil
}

func (c *Coffer) copyMap(inMap map[string][]byte) (map[string][]byte, error) { // на случай, если мы хотим скопировать входные данные запроса, боясь их изменения
	outMap := make(map[string][]byte, len(inMap))
	for k, v := range inMap {
		list := make([]byte, len(v))
		n := copy(list, v)
		if n != len(v) {
			return nil, fmt.Errorf("Slice (bytes) copy failed.")
		}
		outMap[k] = list
	}
	return outMap, nil
}

func (c *Coffer) checkPanic() {
	if err := recover(); err != nil {
		c.hasp.Block()
		//atomic.StoreInt64(&c.hasp, statePanic)
		//fmt.Println(err)
		c.logger.Error(err).Context("Object", "Coffer").Context("Method", "checkPanic").Send()
	}
}

func (c *Coffer) alarmFunc(err error) { // для журнала
	c.logger.Error(err).Context("Object", "Journal").Context("Method", "Write").Send()
}
