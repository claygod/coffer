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

func (c *Coffer) Write(key string, value []byte) error {
	return c.WriteList(map[string][]byte{key: value})
}

func (c *Coffer) WriteListSafe(input map[string][]byte) error { // A method with little protection against changing arguments. Slower.
	inCopy, err := c.copyMap(input)
	if err != nil {
		return err
	}
	return c.WriteList(inCopy)
}

func (c *Coffer) WriteList(input map[string][]byte) error {
	//defer c.checkPanic()
	if !c.hasp.Add() {
		return fmt.Errorf("Coffer is stopped")
	}
	defer c.hasp.Done()
	if ln := len(input); ln > c.config.MaxRecsPerOperation { // контроль максимально допустимого количества добавленных записей за одну операцию
		return fmt.Errorf("The allowable number of entries in operation %d, and in the request %d.", c.config.MaxRecsPerOperation, ln)
	}
	for key, value := range input {
		if ln := len(key); ln > c.config.UsecasesConfig.MaxKeyLength { // контроль максимально допустимой длины ключа
			return fmt.Errorf("The admissible key length is %d; there is a key with a length of %d in the request.", c.config.UsecasesConfig.MaxKeyLength, ln)
		}
		if ln := len(value); ln > c.config.UsecasesConfig.MaxValueLength { // контроль максимально допустимой длины значения
			return fmt.Errorf("The admissible value length is %d; there is a value with a length of %d in the request.", c.config.UsecasesConfig.MaxValueLength, ln)
		}
	}
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
	//TODO: если интерактор возвращает ошибку, возможно нужно всё остановить
	err, wrn := c.recInteractor.WriteList(req)
	if err != nil {
		defer c.Stop()
		return err
	}
	return wrn // при варнинге возвращаем ошибку, но приложение не останавливаем, т.е. это единичный случай и следующий может быть положительным.
}

func (c *Coffer) Read(key string) *reports.ReportRead {
	rep := &reports.ReportRead{}
	repList := c.ReadList([]string{key})
	rep.Code = repList.Code
	rep.Error = repList.Error
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

// func (c *Coffer) ReadList(keys []string) (map[string][]byte, []string, error) {
// 	//defer c.checkPanic()
// 	if !c.hasp.Add() {
// 		return nil, nil, fmt.Errorf("Coffer is stopped")
// 	}
// 	defer c.hasp.Done()
// 	if ln := len(keys); ln > c.config.MaxRecsPerOperation { // контроль максимально допустимого количества добавленных записей за одну операцию
// 		return nil, nil, fmt.Errorf("The allowable number of entries in operation %d, and in the request %d.", c.config.MaxRecsPerOperation, ln)
// 	}
// 	for _, key := range keys { // контроль максимально допустимой длины ключа
// 		if ln := len(key); ln > c.config.UsecasesConfig.MaxKeyLength {
// 			return nil, nil, fmt.Errorf("The admissible key length is %d; there is a key with a length of %d in the request.", c.config.UsecasesConfig.MaxKeyLength, ln)
// 		}
// 	}
// 	c.porter.Catch(keys)
// 	defer c.porter.Throw(keys)
// 	req := &usecases.ReqLoadList{
// 		Time: time.Now(),
// 		Keys: keys,
// 	}
// 	return c.recInteractor.ReadList(req)
//}

func (c *Coffer) ReadList(keys []string) *reports.ReportReadList {
	rep := &reports.ReportReadList{}
	//defer c.checkPanic()
	if !c.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("Coffer is stopped")
		return rep //  nil, nil, fmt.Errorf("Coffer is stopped")
	}
	defer c.hasp.Done()
	//TODO: контроль длин и т.д. должен быть и в других экшенах
	if ln := len(keys); ln > c.config.MaxRecsPerOperation { // контроль максимально допустимого количества добавленных записей за одну операцию
		rep.Code = codes.ErrRecordLimitExceeded
		rep.Error = fmt.Errorf("The allowable number of entries in operation %d, and in the request %d.", c.config.MaxRecsPerOperation, ln)
		return rep //nil, nil, fmt.Errorf("The allowable number of entries in operation %d, and in the request %d.", c.config.MaxRecsPerOperation, ln)
	}
	for _, key := range keys { // контроль максимально допустимой и нулевой длины ключа
		ln := len(key)
		if ln > c.config.UsecasesConfig.MaxKeyLength {
			rep.Code = codes.ErrExceedingMaxKeyLength
			rep.Error = fmt.Errorf("The admissible key length is %d; there is a key with a length of %d in the request.", c.config.UsecasesConfig.MaxKeyLength, ln)
			return rep //nil, nil, fmt.Errorf("The admissible key length is %d; there is a key with a length of %d in the request.", c.config.UsecasesConfig.MaxKeyLength, ln)
		}
		if ln == 0 {
			rep.Code = codes.ErrExceedingZeroKeyLength
			rep.Error = fmt.Errorf("The key length is 0.")
			return rep //nil, nil, fmt.Errorf("The admissible key length is %d; there is a key with a length of %d in the request.", c.config.UsecasesConfig.MaxKeyLength, ln)
		}
	}
	c.porter.Catch(keys)
	defer c.porter.Throw(keys)
	req := &usecases.ReqLoadList{
		Time: time.Now(),
		Keys: keys,
	}
	return c.recInteractor.ReadList(req)
}

func (c *Coffer) Delete(key string) *reports.ReportDelete {
	repList := c.DeleteListStrict([]string{key})
	rep := &reports.ReportDelete{
		Code:  repList.Code,
		Error: repList.Error,
	}
	return rep
}

// func (c *Coffer) DeleteListSafe(keys []string) error { // A method with little protection against changing arguments. Slower.
// 	keysCopy, err := c.copySlice(keys)
// 	if err != nil {
// 		return err
// 	}
// 	return c.DeleteList(keysCopy)
// }

func (c *Coffer) DeleteListStrict(keys []string) *reports.ReportDeleteList {

	//defer c.checkPanic()
	if !c.hasp.Add() {
		rep := &reports.ReportDeleteList{
			Code:  codes.PanicStopped,
			Error: fmt.Errorf("Coffer is stopped"),
		}
		return rep
	}
	defer c.hasp.Done()
	c.porter.Catch(keys)
	defer c.porter.Throw(keys)
	req := &usecases.ReqDeleteList{
		Time: time.Now(),
		Keys: keys,
	}
	rep := c.recInteractor.DeleteListStrict(req)
	if rep.Code >= codes.Panic {
		defer c.Stop()
	}
	return rep
}

func (c *Coffer) TransactionSafe(handlerName string, keys []string, arg interface{}) error { // A method with little protection against changing arguments. Slower.
	keysCopy, err := c.copySlice(keys)
	if err != nil {
		return err
	}
	return c.Transaction(handlerName, keysCopy, arg)
}

func (c *Coffer) Transaction(handlerName string, keys []string, arg interface{}) error {
	//defer c.checkPanic()
	if !c.hasp.Add() {
		return fmt.Errorf("Coffer is stopped")
	}
	defer c.hasp.Done()
	if ln := len(keys); ln > c.config.MaxRecsPerOperation { // контроль максимально допустимого количества добавленных записей за одну операцию
		return fmt.Errorf("The allowable number of entries in operation %d, and in the request %d.", c.config.MaxRecsPerOperation, ln)
	}
	for _, key := range keys { // контроль максимально допустимой длины ключа
		if ln := len(key); ln > c.config.UsecasesConfig.MaxKeyLength {
			return fmt.Errorf("The admissible key length is %d; there is a key with a length of %d in the request.", c.config.UsecasesConfig.MaxKeyLength, ln)
		}
	}
	c.porter.Catch(keys)
	defer c.porter.Throw(keys)

	req := &usecases.ReqTransaction{
		Time:        time.Now(),
		HandlerName: handlerName,
		Keys:        keys,
		Value:       arg,
	}
	err, wrn := c.recInteractor.Transaction(req)
	if err != nil {
		defer c.Stop()
		return err //TODO: возвращать структуру отчёта а не ошибку
	}
	return wrn
}
