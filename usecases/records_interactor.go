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

	"github.com/claygod/coffer/domain"
	"github.com/claygod/coffer/reports"
	"github.com/claygod/coffer/reports/codes"
)

/*
RecordsInteractor - the main request handler for operations with database records.
*/
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
	journal    Journaler
	filenamer  FileNamer
	hasp       Starter
}

/*
NewRecordsInteractor - create new RecordsInteractor
*/
func NewRecordsInteractor(
	config *Config,
	logger Logger,
	loader *Loader,
	chp *checkpoint,
	trs *Transaction,
	reqCoder *ReqCoder,
	repo domain.RecordsRepository,
	handlers domain.HandlersRepository,
	resControl Resourcer,
	journal Journaler,
	filenamer FileNamer,
	hasp Starter) (*RecordsInteractor, error, error) {

	r := &RecordsInteractor{
		config:     config,
		logger:     logger,
		loader:     loader,
		chp:        chp,
		opr:        NewOperations(config, reqCoder, resControl, trs),
		trs:        trs,
		coder:      reqCoder,
		repo:       repo,
		handlers:   handlers,
		resControl: resControl,
		journal:    journal,
		filenamer:  filenamer,
		hasp:       hasp,
	}

	chpList, err := r.filenamer.GetHalf("-1"+extCheck+extPoint, true)
	if err != nil {
		return nil, err, nil
	}

	fChName, err := r.loader.LoadLatestValidCheckpoint(chpList, r.repo) // download the latest valid version of checkpoint
	if err != nil {
		r.logger.Warning(err)
		fChName = "-1" + extCheck + extPoint
	}

	// download all available subsequent logs
	logsList, err := r.filenamer.GetHalf(strings.Replace(fChName, extCheck+extPoint, extLog, -1), true)
	if err != nil {
		return nil, err, nil
	}

	// execute all available subsequent logs
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

/*
Start - start the interactor.
*/
func (r *RecordsInteractor) Start() bool {
	if err := r.journal.Start(); err != nil {
		return false
	}

	return r.hasp.Start()
}

/*
Stop - stop the interactor.
*/
func (r *RecordsInteractor) Stop() bool {
	if !r.hasp.Block() {
		return false
	}

	defer r.hasp.Unblock()
	r.journal.Stop()

	if err := r.save(); err != nil {
		r.logger.Error(err, "Method=Stop")

		return false
	} else if r.config.RemoveUnlessLogs {
		//TODO: here you can delete all junk except the last checkpoint
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
		}
		novName = nm
	}

	novName = strings.Replace(novName, extCheck+extPoint, extCheck, 1)

	if err := r.chp.save(r.repo, novName); err != nil {
		return err
	}

	return nil
}

/*
WriteListOptional - set a few records in safe mode.
*/
func (r *RecordsInteractor) WriteListOptional(req *ReqWriteList) *reports.ReportWriteList {
	rep := &reports.ReportWriteList{Report: reports.Report{}}

	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")

		return rep
	}

	defer r.hasp.Done()

	// prepare the byte version of the operation for the log
	opBytes, err := r.reqWriteListToLog(req)
	if err != nil {
		rep.Code = codes.ErrParseRequest
		rep.Error = err

		return rep
	}

	// check if there are enough resources (memory, disk) to complete the task
	if !r.resControl.GetPermission(int64(len(opBytes))) {
		rep.Code = codes.ErrResources
		rep.Error = fmt.Errorf("Insufficient resources (memory, disk)")

		return rep
	}

	// execute
	rep.Found = r.repo.WriteListOptional(req.List)

	if err := r.journal.Write(opBytes); err != nil {
		defer r.hasp.Stop()
		rep.Code = codes.PanicWAL
		rep.Error = err

		return rep
	}

	rep.Code = codes.Ok

	return rep
}

/*
WriteListStrict - set a few records in strict mode.
*/
func (r *RecordsInteractor) WriteListStrict(req *ReqWriteList) *reports.ReportWriteList {
	rep := &reports.ReportWriteList{Report: reports.Report{}}

	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")

		return rep
	}

	defer r.hasp.Done()

	// prepare the byte version of the operation for the log
	opBytes, err := r.reqWriteListToLog(req)
	if err != nil {
		rep.Code = codes.ErrParseRequest
		rep.Error = err

		return rep
	}

	// check if there are enough resources (memory, disk) to complete the task
	if !r.resControl.GetPermission(int64(len(opBytes))) {
		rep.Code = codes.ErrResources
		rep.Error = fmt.Errorf("Insufficient resources (memory, disk)")

		return rep
	}

	// execute
	rep.Found = r.repo.WriteListStrict(req.List)
	if err := r.journal.Write(opBytes); err != nil {
		defer r.hasp.Stop()
		rep.Code = codes.PanicWAL
		rep.Error = err

		return rep
	}

	if len(rep.Found) == 0 {
		rep.Code = codes.Ok
	} else {
		rep.Code = codes.ErrRecordsFound
	}

	return rep
}

/*
WriteListUnsafe - set a few records in unsafe mode.
*/
func (r *RecordsInteractor) WriteListUnsafe(req *ReqWriteList) *reports.Report {
	rep := &reports.Report{}

	// prepare the byte version of the operation for the log
	opBytes, err := r.reqWriteListToLog(req)

	if err != nil {
		rep.Code = codes.ErrParseRequest
		rep.Error = err

		return rep
	}

	// execute
	r.repo.WriteListUnsafe(req.List)

	if err := r.journal.Write(opBytes); err != nil {
		defer r.hasp.Stop()
		rep.Code = codes.PanicWAL
		rep.Error = err

		return rep
	}

	rep.Code = codes.Ok

	return rep
}

/*
ReadList - get a few records in safe mode.
*/
func (r *RecordsInteractor) ReadList(req *ReqLoadList) *reports.ReportReadList {
	rep := &reports.ReportReadList{Report: reports.Report{}}
	defer r.checkPanic()

	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")

		return rep
	}

	defer r.hasp.Done()
	// execute
	data, notFound := r.repo.ReadList(req.Keys)

	if len(notFound) != 0 {
		rep.Code = codes.ErrReadRecords
	}

	rep.Data = data
	rep.NotFound = notFound

	return rep
}

/*
ReadListUnsafe - get a few records in unsafe mode.
*/
func (r *RecordsInteractor) ReadListUnsafe(req *ReqLoadList) *reports.ReportReadList {
	rep := &reports.ReportReadList{Report: reports.Report{}}
	defer r.checkPanic()

	// выполняем
	data, notFound := r.repo.ReadListUnsafe(req.Keys)

	if len(notFound) != 0 {
		rep.Code = codes.ErrReadRecords
	}

	rep.Data = data
	rep.NotFound = notFound

	return rep
}

/*
DeleteList - delete multiple records in the database.
*/
func (r *RecordsInteractor) DeleteList(req *ReqDeleteList, strictMode bool) *reports.ReportDeleteList {
	defer r.checkPanic()

	rep := &reports.ReportDeleteList{Report: reports.Report{}}
	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")

		return rep
	}

	defer r.hasp.Done()

	// prepare the byte version of the operation for the log
	opBytes, err := r.reqDeleteListToLog(req)
	if err != nil {
		rep.Code = codes.ErrParseRequest
		rep.Error = err

		return rep
	}

	// check if there are enough resources (memory, disk) to complete the task
	if !r.resControl.GetPermission(int64(len(opBytes))) {
		rep.Code = codes.ErrResources
		rep.Error = fmt.Errorf("Insufficient resources (memory, disk)")

		return rep
	}

	// execute
	if strictMode {
		rep = r.deleteListStrict(req.Keys, opBytes)
	} else {
		rep = r.deleteListOptional(req.Keys, opBytes)
	}

	if rep.Code >= codes.Panic {
		defer r.hasp.Stop()
	}

	return rep
}

func (r *RecordsInteractor) deleteListStrict(keys []string, opBytes []byte) *reports.ReportDeleteList {
	rep := &reports.ReportDeleteList{Report: reports.Report{}}

	// execute
	notFound := r.repo.DelListStrict(keys) // during warning - not everything was deleted (some keys are not in the database)
	rep.NotFound = notFound

	if len(notFound) != 0 {
		rep.Code = codes.ErrNotFound
		rep.Error = fmt.Errorf("Keys not found: %s", strings.Join(notFound, ", "))

		return rep
	}

	if err := r.journal.Write(opBytes); err != nil {
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

	// execute
	removedList, notFound := r.repo.DelListOptional(keys) // during warning - not everything was deleted (some keys are not in the database)
	rep.Removed = removedList
	rep.NotFound = notFound

	if err := r.journal.Write(opBytes); err != nil {
		rep.Code = codes.PanicWAL
		rep.Error = err

		return rep
	}

	rep.Code = codes.Ok

	return rep
}

func (r *RecordsInteractor) reqWriteListToLog(req *ReqWriteList) ([]byte, error) {
	// req marshall to bytes
	reqBytes, err := r.coder.ReqWriteListEncode(req)

	if err != nil {
		return nil, err
	}

	// form the operation
	op := &domain.Operation{
		Code: codeWriteList,
		Body: reqBytes,
	}

	// marshall operation in bytes
	return r.opr.operatToLog(op)
}

func (r *RecordsInteractor) reqDeleteListToLog(req *ReqDeleteList) ([]byte, error) {
	// req marshall to bytes
	reqBytes, err := r.coder.ReqDeleteListEncode(req)

	if err != nil {
		return nil, err
	}

	// form the operation
	op := &domain.Operation{
		Code: codeWriteList,
		Body: reqBytes,
	}

	// marshall operation in bytes
	return r.opr.operatToLog(op)
}

func (r *RecordsInteractor) reqTransactionToLog(req *ReqTransaction) ([]byte, error) {
	// req marshall to bytes
	reqBytes, err := r.coder.ReqTransactionEncode(req)

	if err != nil {
		return nil, err
	}

	// form the operation
	op := &domain.Operation{
		Code: codeTransaction,
		Body: reqBytes,
	}

	// marshall operation in bytes
	return r.opr.operatToLog(op)
}

/*
Transaction - complete a transaction.
*/
func (r *RecordsInteractor) Transaction(req *ReqTransaction) *reports.ReportTransaction { // interface{}, map[string][]byte, *domain.Handler
	//tStart := time.Now().UnixNano()
	//defer fmt.Println("Operation time ", time.Now().UnixNano()-tStart)

	rep := &reports.ReportTransaction{Report: reports.Report{}}

	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")

		return rep
	}

	defer r.hasp.Done()

	// prepare the byte version of the operation for the log
	opBytes, err := r.reqTransactionToLog(req)

	if err != nil {
		rep.Code = codes.ErrParseRequest
		rep.Error = err

		return rep
	}

	// check if there are enough resources (memory, disk) to complete the task
	if !r.resControl.GetPermission(int64(len(opBytes))) {
		rep.Code = codes.ErrResources
		rep.Error = fmt.Errorf("Insufficient resources (memory, disk)")

		return rep
	}

	// execute a transaction
	rep = r.trs.doOperationTransaction(req, r.repo)

	if rep.Code >= codes.Panic {
		defer r.hasp.Stop()
	}

	if rep.Code >= codes.Error {
		return rep
	}

	// записываем результат
	if err := r.journal.Write(opBytes); err != nil {
		defer r.hasp.Stop()
		rep.Code = codes.PanicWAL
		rep.Error = err

		return rep
	}

	rep.Code = codes.Ok

	return rep
}

/*
RecordsCount - get the total number of records in the database.
*/
func (r *RecordsInteractor) RecordsCount() *reports.ReportRecordsCount {
	rep := &reports.ReportRecordsCount{Report: reports.Report{}}

	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")

		return rep
	}

	defer r.hasp.Done()

	// execute
	rep.Count = r.repo.CountRecords()
	rep.Code = codes.Ok

	return rep
}

/*
RecordsList - get records list
*/
func (r *RecordsInteractor) RecordsList() *reports.ReportRecordsList {
	rep := &reports.ReportRecordsList{Report: reports.Report{}}

	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")

		return rep
	}

	defer r.hasp.Done()

	// execute
	rep.Data = r.repo.RecordsList()
	rep.Code = codes.Ok

	return rep
}

/*
RecordsListWithPrefix - get records list with prefix.
*/
func (r *RecordsInteractor) RecordsListWithPrefix(prefix string) *reports.ReportRecordsList {
	rep := &reports.ReportRecordsList{Report: reports.Report{}}

	if !r.hasp.Add() {
		rep.Code = codes.PanicStopped
		rep.Error = fmt.Errorf("RecordsInteractor is stopped")

		return rep
	}

	defer r.hasp.Done()

	// execute
	rep.Data = r.repo.RecordsListWithPrefix(prefix)
	rep.Code = codes.Ok

	return rep
}

func (r *RecordsInteractor) findExtraKeys(writeList map[string][]byte, curMap map[string][]byte) error {
	extraKeys := make([]string, 0, len(writeList))

	for key := range writeList {
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

	for key := range arr {
		keys = append(keys, key)
	}

	return keys
}
