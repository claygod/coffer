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

func (c *Coffer) WriteList(input map[string][]byte) *reports.Report {
	rep := &reports.Report{}
	defer c.panicRecover()
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

func (c *Coffer) WriteListUnsafe(input map[string][]byte) *reports.Report {
	rep := &reports.Report{}
	defer c.panicRecover()
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
	req := &usecases.ReqWriteList{
		Time: time.Now(),
		List: input,
	}
	rep = c.recInteractor.WriteListUnsafe(req)
	if rep.Code >= codes.Panic {
		defer c.Stop()
	}
	return rep
}

func (c *Coffer) Read(key string) *reports.ReportRead {
	rep := &reports.ReportRead{Report: reports.Report{}}
	defer c.panicRecover()
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

func (c *Coffer) ReadList(keys []string) *reports.ReportReadList {
	rep := &reports.ReportReadList{Report: reports.Report{}}
	defer c.panicRecover()
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

func (c *Coffer) ReadListUnsafe(keys []string) *reports.ReportReadList {
	rep := &reports.ReportReadList{Report: reports.Report{}}
	defer c.panicRecover()

	if code, err := c.checkLenCountKeys(keys); code != codes.Ok {
		rep.Code = code
		rep.Error = err
		return rep
	}

	req := &usecases.ReqLoadList{
		Time: time.Now(),
		Keys: keys,
	}
	rep = c.recInteractor.ReadListUnsafe(req)
	if rep.Code >= codes.Panic {
		defer c.Stop()
	}
	return rep
}

func (c *Coffer) Delete(key string) *reports.Report {
	repList := c.DeleteListStrict([]string{key})
	return &repList.Report
}

func (c *Coffer) DeleteListStrict(keys []string) *reports.ReportDeleteList {
	return c.deleteList(keys, true)
}

func (c *Coffer) DeleteListOptional(keys []string) *reports.ReportDeleteList {
	return c.deleteList(keys, false)
}

func (c *Coffer) deleteList(keys []string, strictMode bool) *reports.ReportDeleteList {
	rep := &reports.ReportDeleteList{Report: reports.Report{}}
	defer c.panicRecover()
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

func (c *Coffer) Transaction(handlerName string, keys []string, arg []byte) *reports.Report {
	// tStart := time.Now().UnixNano()
	// defer fmt.Println("Время проведения оперерации ", time.Now().UnixNano()-tStart)

	rep := &reports.Report{}
	defer c.panicRecover()
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

func (c *Coffer) Count() *reports.ReportRecordsCount {
	rep := &reports.ReportRecordsCount{Report: reports.Report{}}
	defer c.panicRecover()
	if !c.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("Coffer is stopped")
		return rep
	}
	defer c.hasp.Done()

	rep = c.recInteractor.RecordsCount()
	if rep.Code >= codes.Panic {
		defer c.Stop()
	}
	return rep
}

func (c *Coffer) RecordsList() *reports.ReportRecordsList {
	defer c.panicRecover()
	if !c.hasp.Add() {
		rep := &reports.ReportRecordsList{Report: reports.Report{}}
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("Coffer is stopped")
		return rep
	}
	defer c.hasp.Done()

	rep := c.recInteractor.RecordsList()
	if rep.Code >= codes.Panic {
		defer c.Stop()
	}
	return rep
}

func (c *Coffer) RecordsListWithPrefix(prefix string) *reports.ReportRecordsList {
	defer c.panicRecover()
	if !c.hasp.Add() {
		rep := &reports.ReportRecordsList{Report: reports.Report{}}
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("Coffer is stopped")
		return rep
	}
	defer c.hasp.Done()

	rep := c.recInteractor.RecordsListWithPrefix(prefix)
	if rep.Code >= codes.Panic {
		defer c.Stop()
	}
	return rep
}

func (c *Coffer) RecordsListWithSuffix(suffix string) *reports.ReportRecordsList {
	defer c.panicRecover()
	if !c.hasp.Add() {
		rep := &reports.ReportRecordsList{Report: reports.Report{}}
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("Coffer is stopped")
		return rep
	}
	defer c.hasp.Done()

	rep := c.recInteractor.RecordsListWithSuffix(suffix)
	if rep.Code >= codes.Panic {
		defer c.Stop()
	}
	return rep
}

func (c *Coffer) RecordsListUnsafe() *reports.ReportRecordsList {
	defer c.panicRecover()
	rep := c.recInteractor.RecordsList()
	if rep.Code >= codes.Panic {
		defer c.Stop()
	}
	return rep
}
