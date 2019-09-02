package coffer

// Coffer
// Actions
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	//"strings"
	"time"

	"github.com/claygod/coffer/reports"
	"github.com/claygod/coffer/reports/codes"
	"github.com/claygod/coffer/usecases"
)

func (c *Coffer) Write(key string, value []byte) *reports.Report {
	return c.WriteList(map[string][]byte{key: value})
}

// func (c *Coffer) WriteListSafe(input map[string][]byte) error { // A method with little protection against changing arguments. Slower.
// 	inCopy, err := c.copyMap(input)
// 	if err != nil {
// 		return err
// 	}
// 	return c.WriteList(inCopy)
// }

func (c *Coffer) WriteList(input map[string][]byte) *reports.Report {
	rep := &reports.Report{}
	//defer c.checkPanic()
	if !c.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("Coffer is stopped")
		return rep
	}
	defer c.hasp.Done()
	for _, value := range input {
		if ln := len(value); ln > c.config.UsecasesConfig.MaxValueLength { // контроль максимально допустимой длины значения
			rep.Code = codes.ErrExceedingMaxValueSize
			rep.Error = fmt.Errorf("The admissible value length is %d; there is a value with a length of %d in the request.", c.config.UsecasesConfig.MaxValueLength, ln)
			return rep
		}
	}
	keys := c.extractKeysFromMap(input)
	if code, err := c.checkLenCountKeys(keys); code != codes.Ok {
		rep.Code = code
		rep.Error = err
		return rep
	}

	c.porter.Catch(keys)
	defer c.porter.Throw(keys)
	req := &usecases.ReqWriteList{
		Time: time.Now(), //т.к. время берём ПОСЛЕ операции Catch для конкретно этих записей временных коллизий не будет
		List: input,
	}
	rep = c.recInteractor.WriteList(req)
	if rep.Code >= codes.Panic {
		defer c.Stop()
	}
	return rep
}

func (c *Coffer) Read(key string) *reports.ReportRead {
	rep := &reports.ReportRead{Report: reports.Report{}}
	repList := c.ReadList([]string{key})
	rep.Report = repList.Report
	//rep.Code = repList.Code
	//rep.Error = repList.Error
	if len(repList.Data) == 1 {
		if d, ok := repList.Data[key]; ok {
			rep.Data = d
		}
	}
	return rep
}

// func (c *Coffer) ReadListSafe(keys []string) *reports.ReportReadList { // A method with little protection against changing arguments. Slower.
// 	keysCopy, err := c.copySlice(keys)
// 	if err != nil {

// 		return nil, nil, err
// 	}
// 	return c.ReadList(keysCopy)
// }

func (c *Coffer) ReadList(keys []string) *reports.ReportReadList {
	rep := &reports.ReportReadList{Report: reports.Report{}}
	//defer c.checkPanic()
	if !c.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("Coffer is stopped")
		return rep
	}
	defer c.hasp.Done()

	if code, err := c.checkLenCountKeys(keys); code != codes.Ok { //TODO: контроль длин и т.д. должен быть и в других экшенах
		rep.Code = code
		rep.Error = err
		return rep
	}

	c.porter.Catch(keys)
	defer c.porter.Throw(keys)
	req := &usecases.ReqLoadList{
		Time: time.Now(),
		Keys: keys,
	}
	rep = c.recInteractor.ReadList(req)
	if rep.Code >= codes.Panic {
		defer c.Stop()
	}
	return rep
}

func (c *Coffer) Delete(key string) *reports.Report {
	repList := c.DeleteListStrict([]string{key})
	return &repList.Report
	// rep := &reports.Report{
	// 	Code:  repList.Code,
	// 	Error: repList.Error,
	// }
	// return rep
}

// func (c *Coffer) DeleteListSafe(keys []string) error { // A method with little protection against changing arguments. Slower.
// 	keysCopy, err := c.copySlice(keys)
// 	if err != nil {
// 		return err
// 	}
// 	return c.DeleteList(keysCopy)
// }

func (c *Coffer) DeleteListStrict(keys []string) *reports.ReportDeleteList {
	return c.deleteList(keys, true)
}

func (c *Coffer) DeleteListOptional(keys []string) *reports.ReportDeleteList {
	return c.deleteList(keys, false)
}

func (c *Coffer) deleteList(keys []string, strictMode bool) *reports.ReportDeleteList {
	rep := &reports.ReportDeleteList{Report: reports.Report{}}
	//defer c.checkPanic()
	if !c.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("Coffer is stopped")
		return rep
	}
	defer c.hasp.Done()

	if code, err := c.checkLenCountKeys(keys); code != codes.Ok {
		rep.Code = code
		rep.Error = err
		return rep
	}

	c.porter.Catch(keys)
	defer c.porter.Throw(keys)
	req := &usecases.ReqDeleteList{
		Time: time.Now(),
		Keys: keys,
	}
	rep = c.recInteractor.DeleteList(req, strictMode)
	if rep.Code >= codes.Panic {
		defer c.Stop()
	}
	return rep
}

// func (c *Coffer) TransactionSafe(handlerName string, keys []string, arg interface{}) error { // A method with little protection against changing arguments. Slower.
// 	keysCopy, err := c.copySlice(keys)
// 	if err != nil {
// 		return err
// 	}
// 	return c.Transaction(handlerName, keysCopy, arg)
// }

func (c *Coffer) Transaction(handlerName string, keys []string, arg interface{}) *reports.Report {
	rep := &reports.Report{}
	//defer c.checkPanic()
	if !c.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("Coffer is stopped")
		return rep
	}
	defer c.hasp.Done()

	if code, err := c.checkLenCountKeys(keys); code != codes.Ok {
		rep.Code = code
		rep.Error = err
		return rep
	}

	c.porter.Catch(keys)
	defer c.porter.Throw(keys)

	req := &usecases.ReqTransaction{
		Time:        time.Now(),
		HandlerName: handlerName,
		Keys:        keys,
		Value:       arg,
	}
	rep = c.recInteractor.Transaction(req)
	if rep.Code >= codes.Panic {
		defer c.Stop()
	}
	return rep
}
