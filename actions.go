package coffer

// Coffer
// Actions
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"time"

	"github.com/claygod/coffer/usecases"
)

func (c *Coffer) Write(key string, value []byte) error {
	if !c.hasp.Add() {
		return fmt.Errorf("Coffer is stopped")
	}
	defer c.hasp.Done()
	return c.WriteList(map[string][]byte{key: value})
}

func (c *Coffer) WriteList(input map[string][]byte) error {
	//TODO: контроль максимально допустимого количества добавленных записей за одну операцию
	defer c.checkPanic()
	keysList := make([]string, 0, len(input))
	for key, _ := range input {
		keysList = append(keysList, key)
	}
	c.porter.Catch(keysList)
	defer c.porter.Throw(keysList)
	req := &usecases.ReqWriteList{
		Time: time.Now(), //т.к. время берём ПОСЛЕ операции Catch для конкретно этих записей временных коллизий не будет
		List: input,
	}
	return c.recInteractor.WriteList(req)
}

func (c *Coffer) Read(key string) ([]byte, error) {
	recs, err := c.ReadList([]string{key})
	if err != nil {
		return nil, err
	}
	rec, ok := recs[key]
	if !ok {
		return nil, fmt.Errorf("Record `%s` not found", key)
	}
	return rec, nil
}

func (c *Coffer) ReadList(keys []string) (map[string][]byte, error) {
	//TODO: контроль максимально допустимого количества чтения записей за одну операцию
	defer c.checkPanic()
	c.porter.Catch(keys)
	defer c.porter.Throw(keys)
	req := &usecases.ReqLoadList{
		Time: time.Now(),
		Keys: keys,
	}
	return c.recInteractor.ReadList(req)
}

func (c *Coffer) Delete(key string) error {
	return c.DeleteList([]string{key})
}

func (c *Coffer) DeleteList(keys []string) error {
	defer c.checkPanic()
	c.porter.Catch(keys)
	defer c.porter.Throw(keys)
	req := &usecases.ReqDeleteList{
		Time: time.Now(),
		Keys: keys,
	}
	return c.recInteractor.DeleteList(req)
}

func (c *Coffer) Transaction(handlerName string, keys []string, arg interface{}) error {
	defer c.checkPanic()
	return nil //TODO:
}

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
