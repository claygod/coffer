package coffer

// Coffer
// Helpers
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"

	"github.com/claygod/coffer/reports/codes"
)

// func (c *Coffer) copySlice(inList []string) ([]string, error) { // на случай, если мы хотим скопировать входные данные запроса, боясь их изменения
// 	outList := make([]string, len(inList))
// 	n := copy(outList, inList)
// 	if n != len(inList) {
// 		return nil, fmt.Errorf("Slice (strings) copy failed.")
// 	}
// 	return outList, nil
// }

// func (c *Coffer) copyMap(inMap map[string][]byte) (map[string][]byte, error) { // на случай, если мы хотим скопировать входные данные запроса, боясь их изменения
// 	outMap := make(map[string][]byte, len(inMap))
// 	for k, v := range inMap {
// 		list := make([]byte, len(v))
// 		n := copy(list, v)
// 		if n != len(v) {
// 			return nil, fmt.Errorf("Slice (bytes) copy failed.")
// 		}
// 		outMap[k] = list
// 	}
// 	return outMap, nil
// }

// func (c *Coffer) checkPanic() {
// 	if err := recover(); err != nil {
// 		c.hasp.Block()
// 		//atomic.StoreInt64(&c.hasp, statePanic)
// 		//fmt.Println(err)
// 		c.logger.Error(err).Context("Object", "Coffer").Context("Method", "checkPanic").Send()
// 	}
// }

// func (c *Coffer) alarmFunc(err error) { // для журнала
// 	c.logger.Error(err).Context("Object", "Journal").Context("Method", "Write").Send()
// }

func (c *Coffer) checkLenCountKeys(keys []string) (codes.Code, error) { // checking operation key parameters
	if ln := len(keys); ln > c.config.MaxRecsPerOperation { // control the maximum allowable number of added records per operation
		return codes.ErrRecordLimitExceeded, fmt.Errorf("The allowable number of entries in operation %d, and in the request %d.", c.config.MaxRecsPerOperation, ln)
	}
	for _, key := range keys { // control of the maximum permissible and zero key length
		ln := len(key)
		if ln > c.config.UsecasesConfig.MaxKeyLength {
			return codes.ErrExceedingMaxKeyLength, fmt.Errorf("The admissible key length is %d; there is a key with a length of %d in the request.", c.config.UsecasesConfig.MaxKeyLength, ln)
		}
		if ln == 0 {
			return codes.ErrExceedingZeroKeyLength, fmt.Errorf("The key length is 0.")
		}
	}
	return codes.Ok, nil
}

func (c *Coffer) extractKeysFromMap(input map[string][]byte) []string {
	keys := make([]string, 0, len(input))
	for key, _ := range input {
		keys = append(keys, key)
	}
	return keys
}

// func (c *Coffer) checkPanic() {
// 	if r := recover(); r != nil {
// 		c.logger.Error(r, "Object=Coffer")
// 	}
// }
