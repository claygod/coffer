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

/*
Write a new record in the database, specifying the key and value.
Their length must satisfy the requirements specified in the configuration.
*/
func (c *Coffer) Write(key string, value []byte) *reports.ReportWriteList {
	return c.WriteList(map[string][]byte{key: value}, false)
}

/*
WriteList - write several records to the database by specifying `map` in the arguments.
Strict mode (true):
	The operation will be performed if there are no records with such keys yet.
	Otherwise, a list of existing records is returned.
Optional mode (false):
	The operation will be performed regardless of whether there are records with such keys or not.
	A list of existing records is returned.
Important: this argument is a reference; it cannot be changed in the calling code!
*/
func (c *Coffer) WriteList(input map[string][]byte, strictMode bool) *reports.ReportWriteList {
	rep := &reports.ReportWriteList{Report: reports.Report{}}

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
		Time: time.Now(),
		List: input,
	}

	if strictMode {
		rep = c.recInteractor.WriteListStrict(req)
	} else {
		rep = c.recInteractor.WriteListOptional(req)
	}

	if rep.Code >= codes.Panic {
		defer c.Stop()
	}

	return rep
}

/*
WriteListUnsafe - write several records to the database by specifying `map` in the arguments.
This method exists in order to fill it up a little faster before starting the database.
The method does not imply concurrent use.
*/
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

/*
Read one entry from the database. In the received `report` there will be a result code, and if it is positive,
that will be the value in the `data` field.
*/
func (c *Coffer) Read(key string) *reports.ReportRead {
	rep := &reports.ReportRead{Report: reports.Report{}}
	defer c.panicRecover()
	repList := c.ReadList([]string{key})
	rep.Report = repList.Report

	if len(repList.Data) == 1 {
		if d, ok := repList.Data[key]; ok {
			rep.Data = d
		}
	}

	return rep
}

/*
ReadList - read a few entries. There is a limit on the maximum number of readable entries in the configuration.
In addition to the found records, a list of not found records is returned.
*/
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

/*
ReadListUnsafe - read a few entries. The method can be called when the database is stopped (not running).
The method does not imply concurrent use.
*/
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

/*
Delete - remove a single record.
*/
func (c *Coffer) Delete(key string) *reports.Report {
	repList := c.DeleteList([]string{key}, true)

	return &repList.Report
}

/*
DeleteList - delete multiple entries.
Delete list Strict - delete several records, but only if they are all in the database.
If at least one entry is missing, then no record will be deleted.
Delete list Optional - delete multiple entries. Those entries from the list
that will be found in the database will be deleted.
*/
func (c *Coffer) DeleteList(keys []string, strictMode bool) *reports.ReportDeleteList {
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

/*
Transaction - execute a handler. The transaction handler must be registered in the database at the stage
of creating and configuring the database. Responsibility for the consistency of the functionality
of transaction handlers between different database launches rests with the database user.
The transaction returns the new values stored in the database.
*/
func (c *Coffer) Transaction(handlerName string, keys []string, arg []byte) *reports.ReportTransaction {
	rep := &reports.ReportTransaction{Report: reports.Report{}}

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

/*
Count - get the number of records in the database.
A query can only be made to a running database
*/
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

/*
CountUnsafe - get the number of records in the database.
Queries to a stopped / not running database cannot be done in parallel!
*/
func (c *Coffer) CountUnsafe() *reports.ReportRecordsCount {
	rep := &reports.ReportRecordsCount{Report: reports.Report{}}

	defer c.panicRecover()

	if !c.hasp.IsReady() && !c.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("Coffer is started, !c.hasp.Add()")

		return rep
	}

	defer c.hasp.Done()

	rep = c.recInteractor.RecordsCount()

	if rep.Code >= codes.Panic {
		defer c.Stop()
	}

	return rep
}

/*
RecordsList - get a list of all database keys. With a large number of records in the database,
the query will be slow, so use its only in case of emergency.
The method only works when the database is running.
*/
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

/*
RecordsListUnsafe - get a list of all database keys. With a large number of records in the database,
the query will be slow, so use its only in case of emergency. When using a query with
a stopped/not_running database, competitiveness prohibited.
*/
func (c *Coffer) RecordsListUnsafe() *reports.ReportRecordsList {
	defer c.panicRecover()

	rep := c.recInteractor.RecordsList()

	if rep.Code >= codes.Panic {
		defer c.Stop()
	}

	return rep
}

/*
RecordsListWithPrefix - get a list of all the keys having prefix
specified in the argument (start with that string).
*/
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
