package usecases

// Coffer
// Records interactor
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"

	"github.com/claygod/coffer/domain"
	//"github.com/claygod/coffer/services/journal"
)

type RecordsInteractor struct {
	// logger Logger
	// chp    *checkpoint
	opr        *operations
	coder      *ReqCoder
	repo       domain.RecordsRepository
	resControl Resourcer
	porter     Porter
	journal    Journaler
	hasp       Starter
}

func (r *RecordsInteractor) Start() bool {
	if !r.hasp.Start() {
		return false
	}
	return true
}

func (r *RecordsInteractor) Stop() bool {
	if !r.hasp.Stop() {
		return false
	}
	return true
}

func (r *RecordsInteractor) WriteList(req *ReqWriteList) error {
	//TODO:
	if !r.hasp.Add() {
		return fmt.Errorf("RecordsInteractor is stopped")
	}
	defer r.hasp.Done()

	// req маршаллим в байты
	reqBytes, err := r.coder.ReqWriteListEncode(req)
	if err != nil {
		return err
	}
	// формируем операцию
	op := &domain.Operation{
		Code: codeWriteList,
		Body: reqBytes,
	}
	// операцию маршаллим в байты
	opBytes, err := r.opr.operatToLog(op)
	if err != nil {
		return err
	}
	// проверяем, достаточно ли ресурсов (памяти, диска) для выполнения задачи
	if r.resControl.GetPermission(int64(len(opBytes))) {
		return fmt.Errorf("Insufficient resources (memory, disk)")
	}
	// блокируем нужные записи
	keys := getKeysFromMap(req.List)
	r.porter.Catch(keys)
	defer r.porter.Throw(keys)
	// проводим операцию  с inmemory хранилищем
	r.repo.WriteList(req.List)
	// журналируем операцию
	r.journal.Write(opBytes)
	return nil
}

func (r *RecordsInteractor) ReadList(req *ReqLoadList) (map[string][]byte, error) {
	//TODO:
	if !r.hasp.Add() {
		return nil, fmt.Errorf("RecordsInteractor is stopped")
	}
	defer r.hasp.Done()

	return nil, nil
}

func (r *RecordsInteractor) DeleteList(req *ReqDeleteList) error {
	//TODO:
	if !r.hasp.Add() {
		return fmt.Errorf("RecordsInteractor is stopped")
	}
	defer r.hasp.Done()

	return nil
}

// func (r *RecordsInteractor) GetRecords([]string) ([]*domain.Record, error) { // (map[string][]byte, error)
// 	return nil, nil
// }

// func (r *RecordsInteractor) SetRecords([]*domain.Record) error { // map[string][]byte
// 	return nil
// }
// func (r *RecordsInteractor) DelRecords([]string) error {
// 	return nil
// }
// func (r *RecordsInteractor) SetUnsafeRecord(*domain.Record) error {
// 	return nil
// }
func (r *RecordsInteractor) Transaction(interface{}, map[string][]byte, *domain.Handler) (map[string][]byte, error) {
	//TODO:
	if !r.hasp.Add() {
		return nil, fmt.Errorf("RecordsInteractor is stopped")
	}
	defer r.hasp.Done()

	return nil, nil
}
