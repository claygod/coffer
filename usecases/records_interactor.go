package usecases

// Coffer
// Records interactor
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"

	//"time"

	"github.com/claygod/coffer/domain"
	"github.com/claygod/coffer/reports"
	"github.com/claygod/coffer/reports/codes"
	//"github.com/claygod/coffer/services/journal"
)

type RecordsInteractor struct {
	config     *Config
	logger     Logger
	loader     *Loader
	chp        *checkpoint
	opr        *Operations
	trs        *Transaction
	coder      *ReqCoder
	repo       domain.RecordsRepository
	handlers   domain.HandlersRepository
	resControl Resourcer
	porter     Porter
	journal    Journaler
	filenamer  FileNamer
	hasp       Starter
}

func NewRecordsInteractor(
	config *Config,
	logger Logger,
	loader *Loader,
	chp *checkpoint,
	//opr *Operations,
	trs *Transaction,
	reqCoder *ReqCoder,
	repo domain.RecordsRepository,
	handlers domain.HandlersRepository,
	resControl Resourcer,
	porter Porter,
	journal Journaler,
	filenamer FileNamer,
	hasp Starter) (*RecordsInteractor, error, error) {

	//oper := NewOperations(logger, config, reqCoder, resControl, trs)

	r := &RecordsInteractor{
		config:     config,
		logger:     logger,
		loader:     loader,
		chp:        chp,
		opr:        NewOperations(logger, config, reqCoder, resControl, trs), //opr,
		trs:        trs,
		coder:      reqCoder,
		repo:       repo,
		handlers:   handlers,
		resControl: resControl,
		porter:     porter,
		journal:    journal,
		filenamer:  filenamer,
		hasp:       hasp,
	}

	chpList, err := r.filenamer.GetHalf("-1"+extCheck+extPoint, true)
	if err != nil {
		return nil, err, nil
	}
	fChName, err := r.loader.LoadLatestValidCheckpoint(chpList, r.repo) // загрузить последнюю валидную версию checkpoint
	if err != nil {
		r.logger.Warning(err)
		fChName = "-1" + extCheck + extPoint
	}

	// загрузить все имеющиеся последующие логи
	logsList, err := r.filenamer.GetHalf(strings.Replace(fChName, extCheck+extPoint, extLog, -1), true) // GetAfterLatest(strings.Replace(fChName, extCheck+extPoint, extLog, -1))
	if err != nil {
		//fmt.Println(22220001, fChName, err)
		return nil, err, nil
	}
	// выполнить все имеющиеся последующие логи
	if len(logsList) > 0 {
		// eсли последний по номеру не `checkpoint`, значит была аварийная остановка,
		// и нужно загрузить всё, что можно, сохранить, и только потом продолжить
		err, wrn := r.loader.LoadLogs(logsList, r.repo)
		if err != nil {
			return nil, err, wrn
		}
		if err := r.save(); err != nil {
			return nil, err, wrn
		}
		r.journal.Restart()
	}

	return r, nil, nil
}

func (r *RecordsInteractor) Start() bool {
	if err := r.journal.Start(); err != nil {
		return false
	}
	return r.hasp.Start()
	// if !r.hasp.Start() {
	// 	return false
	// }
	// return true
}

func (r *RecordsInteractor) Stop() bool {
	if !r.hasp.Block() {
		return false
	}
	defer r.hasp.Unblock()
	r.journal.Stop()
	if err := r.save(); err != nil {
		r.logger.Error(err, "Object:RecordsInteractor", "Method:Stop")
		return false
	} else if r.config.RemoveUnlessLogs {
		//TODO: тут можно удалять всё старьё кроме последнего чекпоинта
	}
	return true
}

func (r *RecordsInteractor) save(args ...string) error {
	var novName string
	if len(args) == 1 {
		novName = args[0]
	} else {
		nm, err := r.filenamer.GetNewFileName(extCheck + extPoint)
		if err != nil {
			return err
		} else {
			novName = nm
		}
	}
	novName = strings.Replace(novName, extCheck+extPoint, extCheck, 1) // r.config.DirPath +
	if err := r.chp.save(r.repo, novName); err != nil {
		return err
	}
	//r.journal.Restart()
	return nil
}

// func (r *RecordsInteractor) WriteList222(req *ReqWriteList) (error, error) {
// 	if !r.hasp.Add() {
// 		return nil, fmt.Errorf("RecordsInteractor is stopped")
// 	}
// 	defer r.hasp.Done()
// 	// подготавливаем байтовую версию операции для лога
// 	opBytes, err := r.reqWriteListToLog(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// проверяем, достаточно ли ресурсов (памяти, диска) для выполнения задачи
// 	if !r.resControl.GetPermission(int64(len(opBytes))) {
// 		return nil, fmt.Errorf("Insufficient resources (memory, disk)")
// 	}
// 	// блокируем нужные записи
// 	keys := r.getKeysFromMap(req.List)
// 	r.porter.Catch(keys)
// 	defer r.porter.Throw(keys)
// 	// выполняем
// 	r.repo.WriteList(req.List)                       // проводим операцию  с inmemory хранилищем
// 	if err := r.journal.Write(opBytes); err != nil { // журналируем операцию
// 		r.hasp.Stop()
// 		return err, nil
// 	}
// 	return nil, nil
// }

func (r *RecordsInteractor) WriteList(req *ReqWriteList) *reports.Report {
	rep := &reports.Report{}
	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")
		return rep
	}
	defer r.hasp.Done()
	// подготавливаем байтовую версию операции для лога
	opBytes, err := r.reqWriteListToLog(req)
	if err != nil {
		rep.Code = codes.ErrParseRequest
		rep.Error = err
		return rep
	}
	// проверяем, достаточно ли ресурсов (памяти, диска) для выполнения задачи
	if !r.resControl.GetPermission(int64(len(opBytes))) {
		rep.Code = codes.ErrResources
		rep.Error = fmt.Errorf("Insufficient resources (memory, disk)")
		return rep
	}
	// выполняем
	r.repo.WriteList(req.List)                       // проводим операцию  с inmemory хранилищем
	if err := r.journal.Write(opBytes); err != nil { // журналируем операцию
		defer r.hasp.Stop()
		rep.Code = codes.PanicWAL
		rep.Error = err
		return rep
	}
	rep.Code = codes.Ok
	return rep
}

func (r *RecordsInteractor) WriteListUnsafe(req *ReqWriteList) *reports.Report {
	rep := &reports.Report{}
	// подготавливаем байтовую версию операции для лога
	opBytes, err := r.reqWriteListToLog(req)
	if err != nil {
		rep.Code = codes.ErrParseRequest
		rep.Error = err
		return rep
	}
	// выполняем
	r.repo.WriteList(req.List)                       // проводим операцию  с inmemory хранилищем
	if err := r.journal.Write(opBytes); err != nil { // журналируем операцию
		defer r.hasp.Stop()
		rep.Code = codes.PanicWAL
		rep.Error = err
		return rep
	}
	rep.Code = codes.Ok
	return rep
}

func (r *RecordsInteractor) ReadList(req *ReqLoadList) *reports.ReportReadList {
	rep := &reports.ReportReadList{Report: reports.Report{}}
	//defer c.checkPanic()
	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")
		return rep //  nil, nil, fmt.Errorf("Coffer is stopped")
	}
	defer r.hasp.Done()
	// выполняем
	data, notFound := r.repo.ReadList(req.Keys)
	if len(notFound) != 0 {
		rep.Code = codes.ErrReadRecords
	}
	//rep.Code = codes.ErrReadRecords
	rep.Data = data
	rep.NotFound = notFound
	return rep
}

func (r *RecordsInteractor) DeleteList(req *ReqDeleteList, strictMode bool) *reports.ReportDeleteList {
	//strictMode := true // strict
	rep := &reports.ReportDeleteList{Report: reports.Report{}}
	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")
		return rep
	}
	defer r.hasp.Done()
	// подготавливаем байтовую версию операции для лога
	opBytes, err := r.reqDeleteListToLog(req)
	if err != nil {
		rep.Code = codes.ErrParseRequest
		rep.Error = err
		return rep
	}
	// проверяем, достаточно ли ресурсов (памяти, диска) для выполнения задачи
	if !r.resControl.GetPermission(int64(len(opBytes))) {
		rep.Code = codes.ErrResources
		rep.Error = fmt.Errorf("Insufficient resources (memory, disk)")
		return rep
	}
	// выполняем
	if strictMode {
		rep = r.deleteListStrict(req.Keys, opBytes)
	} else {
		rep = r.deleteListOptional(req.Keys, opBytes)
	}
	if rep.Code >= codes.Panic {
		defer r.hasp.Stop()
	}

	// notFound := r.repo.DelListStrict(req.Keys) // при варнинге - не всё удалось удалить (каких-то ключей нет в базе)
	// rep.NotFound = notFound
	// if len(notFound) != 0 {
	// 	rep.Code = codes.ErrNotFound
	// 	rep.Error = fmt.Errorf("Keys not found: %s", strings.Join(notFound, ", "))
	// 	return rep
	// }
	// if err := r.journal.Write(opBytes); err != nil { // журналируем операцию
	// 	defer r.hasp.Stop()
	// 	rep.Code = codes.PanicWAL
	// 	rep.Error = err
	// 	return rep
	// }
	// rep.Code = codes.Ok
	// rep.Removed = req.Keys
	return rep
}

func (r *RecordsInteractor) deleteListStrict(keys []string, opBytes []byte) *reports.ReportDeleteList {
	rep := &reports.ReportDeleteList{Report: reports.Report{}}
	// выполняем
	notFound := r.repo.DelListStrict(keys) // при варнинге - не всё удалось удалить (каких-то ключей нет в базе)
	rep.NotFound = notFound
	if len(notFound) != 0 {
		rep.Code = codes.ErrNotFound
		rep.Error = fmt.Errorf("Keys not found: %s", strings.Join(notFound, ", "))
		return rep
	}
	if err := r.journal.Write(opBytes); err != nil { // журналируем операцию
		rep.Code = codes.PanicWAL
		rep.Error = err
		return rep
	}
	rep.Code = codes.Ok
	rep.Removed = keys
	return rep
}

func (r *RecordsInteractor) deleteListOptional(keys []string, opBytes []byte) *reports.ReportDeleteList {
	rep := &reports.ReportDeleteList{Report: reports.Report{}}
	// выполняем
	removedList, notFound := r.repo.DelListOptional(keys) // при варнинге - не всё удалось удалить (каких-то ключей нет в базе)
	rep.Removed = removedList
	rep.NotFound = notFound
	if err := r.journal.Write(opBytes); err != nil { // журналируем операцию
		rep.Code = codes.PanicWAL
		rep.Error = err
		return rep
	}
	rep.Code = codes.Ok
	return rep
}

func (r *RecordsInteractor) reqWriteListToLog(req *ReqWriteList) ([]byte, error) {
	// req маршаллим в байты
	reqBytes, err := r.coder.ReqWriteListEncode(req)
	if err != nil {
		return nil, err
	}
	// формируем операцию
	op := &domain.Operation{
		Code: codeWriteList,
		Body: reqBytes,
	}
	// операцию маршаллим в байты
	return r.opr.operatToLog(op)
}

func (r *RecordsInteractor) reqDeleteListToLog(req *ReqDeleteList) ([]byte, error) {
	// req маршаллим в байты
	reqBytes, err := r.coder.ReqDeleteListEncode(req)
	if err != nil {
		return nil, err
	}
	// формируем операцию
	op := &domain.Operation{
		Code: codeWriteList,
		Body: reqBytes,
	}
	// операцию маршаллим в байты
	return r.opr.operatToLog(op)
}

func (r *RecordsInteractor) reqTransactionToLog(req *ReqTransaction) ([]byte, error) {
	// req маршаллим в байты
	reqBytes, err := r.coder.ReqTransactionEncode(req)
	if err != nil {
		return nil, err
	}
	// формируем операцию
	op := &domain.Operation{
		Code: codeTransaction,
		Body: reqBytes,
	}
	// операцию маршаллим в байты
	return r.opr.operatToLog(op)
}

func (r *RecordsInteractor) Transaction(req *ReqTransaction) *reports.Report { // interface{}, map[string][]byte, *domain.Handler
	//tStart := time.Now().UnixNano()
	//defer fmt.Println("Время проведения оперерации ", time.Now().UnixNano()-tStart)

	rep := &reports.Report{}
	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")
		return rep
	}
	defer r.hasp.Done()
	// // ищем хэндлер
	// hdl, err := r.handlers.Get(req.HandlerName)
	// if err != nil {
	// 	return nil, err
	// }
	// handler := *hdl
	// подготавливаем байтовую версию операции для лога
	opBytes, err := r.reqTransactionToLog(req)
	if err != nil {
		rep.Code = codes.ErrParseRequest
		rep.Error = err
		return rep
	}
	// проверяем, достаточно ли ресурсов (памяти, диска) для выполнения задачи
	if !r.resControl.GetPermission(int64(len(opBytes))) {
		rep.Code = codes.ErrResources
		rep.Error = fmt.Errorf("Insufficient resources (memory, disk)")
		return rep
	}
	// блокируем нужные записи
	r.porter.Catch(req.Keys)
	defer r.porter.Throw(req.Keys)

	// выполняем транзакцию
	rep = r.trs.doOperationTransaction(req, r.repo)
	// находим хандлер
	// читаем текущие значения
	// проводим операцию  с полученными из репо значениями
	// проверяем, нет ли надобности удалить какие-то записи (готовим список)
	// сохранение изменённых записей (полученных в результате выполнения транзакции)
	// удаление записей (при необходимости)
	if rep.Code >= codes.Panic {
		defer r.hasp.Stop()
	}
	if rep.Code >= codes.Error {
		return rep
	}

	// curMap, notFound := r.repo.ReadList(req.Keys)
	// if len(notFound) != 0 {
	// 	return nil, fmt.Errorf("Records not found: %s", strings.Join(notFound, ", "))
	// }
	// // выполняем транзакцию
	// writeList, err := handler(req.Value, curMap)
	// if err != nil {
	// 	return nil, err
	// }
	// // проверяем, нет ли "лишних" ключей в ответе
	// if err := r.findExtraKeys(writeList, curMap); err != nil {
	// 	return nil, err
	// }
	// записываем результат
	// r.repo.WriteList(writeList)                      // проводим операцию  с inmemory хранилищем
	if err := r.journal.Write(opBytes); err != nil { // журналируем операцию
		defer r.hasp.Stop()
		rep.Code = codes.PanicWAL
		rep.Error = err
		return rep
	}
	rep.Code = codes.Ok
	return rep
}

func (r *RecordsInteractor) RecordsCount() *reports.ReportRecordsCount {
	rep := &reports.ReportRecordsCount{Report: reports.Report{}}
	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")
		return rep
	}
	defer r.hasp.Done()
	// выполняем
	rep.Count = r.repo.CountRecords() // проводим операцию  с inmemory хранилищем
	rep.Code = codes.Ok
	return rep
}

func (r *RecordsInteractor) findExtraKeys(writeList map[string][]byte, curMap map[string][]byte) error {
	extraKeys := make([]string, 0, len(writeList))
	for key, _ := range writeList {
		if _, ok := curMap[key]; !ok {
			extraKeys = append(extraKeys, key)
		}
	}
	if len(extraKeys) > 0 {
		return fmt.Errorf("Found extra keys: %s", strings.Join(extraKeys, " , "))
	}
	return nil
}

func (r *RecordsInteractor) findLogsAfterCheckpoint(chpName string) ([]string, error) {
	logBarrier, err := strconv.ParseInt(strings.Replace(chpName, extCheck+extPoint, "", 1), 10, 64)
	if err != nil {
		return nil, err
	}
	logsNames, err := r.getFilesByExtList(extLog)
	if err != nil {
		return nil, err
	}
	for i, logName := range logsNames {
		num, err := strconv.ParseInt(strings.Replace(logName, extLog, "", 1), 10, 64)
		if err != nil {
			return nil, err
		}
		if num > logBarrier {
			return logsNames[i : len(logsNames)-1], nil
		}
	}
	return make([]string, 0), nil
}

func (r *RecordsInteractor) getFilesByExtList(ext string) ([]string, error) {
	files, err := ioutil.ReadDir(r.config.DirPath)
	if err != nil {
		return nil, err
	}
	list := make([]string, 0, len(files))
	for _, fl := range files {
		if strings.HasSuffix(fl.Name(), ext) {
			list = append(list, fl.Name())
		}
	}
	sort.Strings(list)
	return list, nil
}

func (r *RecordsInteractor) getKeysFromMap(arr map[string][]byte) []string {
	keys := make([]string, 0, len(arr))
	for key, _ := range arr {
		keys = append(keys, key)
	}
	return keys
}
