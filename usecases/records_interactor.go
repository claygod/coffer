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
	"time"

	"github.com/claygod/coffer/domain"
	//"github.com/claygod/coffer/services/journal"
)

type RecordsInteractor struct {
	config     *Config
	logger     Logger
	chp        *checkpoint
	opr        *operations
	coder      *ReqCoder
	repo       domain.RecordsRepository
	handlers   domain.HandlersRepository
	resControl Resourcer
	porter     Porter
	journal    Journaler
	//pathDir    string
	hasp Starter
}

func NewRecordsInteractor(
	config *Config,
	logger Logger,
	chp *checkpoint,
	opr *operations,
	coder *ReqCoder,
	repo domain.RecordsRepository,
	handlers domain.HandlersRepository,
	resControl Resourcer,
	porter Porter,
	journal Journaler,
	hasp Starter,
	pathDir string) (*RecordsInteractor, error) {

	r := &RecordsInteractor{
		config:     config,
		logger:     logger,
		chp:        chp,
		opr:        opr,
		coder:      coder,
		repo:       repo,
		handlers:   handlers,
		resControl: resControl,
		porter:     porter,
		journal:    journal,
		hasp:       hasp,
	}

	// загрузить последнюю версию checkpoint
	fChName, err := r.findLatestCheckpoint()
	if err != nil {
		return nil, err
	}
	if err := r.chp.load(r.repo, fChName); err != nil { //загружаем последний checkpoint
		return nil, err
	}
	// загрузить и выполнить все имеющиеся последующие логи
	logsList, err := r.findLogsAfterCheckpoint(fChName)
	if err != nil {
		return nil, err
	}
	if err := r.loadLogs(logsList); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *RecordsInteractor) Start() bool {
	if !r.hasp.Start() {
		return false
	}
	return true
}

func (r *RecordsInteractor) Stop() bool {
	if !r.hasp.Block() {
		return false
	}
	defer r.hasp.Unblock()
	if err := r.save(); err != nil {
		r.logger.Error(err)
		return false
	}
	return true
}

func (r *RecordsInteractor) save() error {
	novName := strconv.FormatInt(time.Now().Unix(), 10) + ".check"
	return r.chp.save(r.repo, novName)
}

func (r *RecordsInteractor) WriteList(req *ReqWriteList) error {
	if !r.hasp.Add() {
		return fmt.Errorf("RecordsInteractor is stopped")
	}
	defer r.hasp.Done()
	// подготавливаем байтовую версию операции для лога
	opBytes, err := r.reqWriteListToLog(req)
	if err != nil {
		return err
	}
	// проверяем, достаточно ли ресурсов (памяти, диска) для выполнения задачи
	if r.resControl.GetPermission(int64(len(opBytes))) {
		return fmt.Errorf("Insufficient resources (memory, disk)")
	}
	// блокируем нужные записи
	keys := r.getKeysFromMap(req.List)
	r.porter.Catch(keys)
	defer r.porter.Throw(keys)
	// выполняем
	r.repo.WriteList(req.List) // проводим операцию  с inmemory хранилищем
	r.journal.Write(opBytes)   // журналируем операцию
	return nil
}

func (r *RecordsInteractor) ReadList(req *ReqLoadList) (map[string][]byte, error) {
	if !r.hasp.Add() {
		return nil, fmt.Errorf("RecordsInteractor is stopped")
	}
	defer r.hasp.Done()
	// блокируем нужные записи
	r.porter.Catch(req.Keys)
	defer r.porter.Throw(req.Keys)
	// выполняем
	return r.repo.ReadList(req.Keys)
}

func (r *RecordsInteractor) DeleteList(req *ReqDeleteList) error {
	if !r.hasp.Add() {
		return fmt.Errorf("RecordsInteractor is stopped")
	}
	defer r.hasp.Done()
	// блокируем нужные записи
	r.porter.Catch(req.Keys)
	defer r.porter.Throw(req.Keys)
	// выполняем
	return r.repo.DelList(req.Keys)
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

func (r *RecordsInteractor) Transaction(req *ReqTransaction) error { // interface{}, map[string][]byte, *domain.Handler
	if !r.hasp.Add() {
		return fmt.Errorf("RecordsInteractor is stopped")
	}
	defer r.hasp.Done()
	// ищем хэндлер
	hdl, err := r.handlers.Get(req.HandlerName)
	if err != nil {
		return err
	}
	handler := *hdl
	// подготавливаем байтовую версию операции для лога
	opBytes, err := r.reqTransactionToLog(req)
	if err != nil {
		return err
	}
	// проверяем, достаточно ли ресурсов (памяти, диска) для выполнения задачи
	if r.resControl.GetPermission(int64(len(opBytes))) {
		return fmt.Errorf("Insufficient resources (memory, disk)")
	}
	// блокируем нужные записи
	r.porter.Catch(req.Keys)
	defer r.porter.Throw(req.Keys)
	// берём текущие значения в записях
	curMap, err := r.repo.ReadList(req.Keys)
	if err != nil {
		return err
	}
	// выполняем транзакцию
	writeList, err := handler(req.Value, curMap)
	if err != nil {
		return err
	}
	// проверяем, нет ли "лишних" ключей в ответе
	if err := r.findExtraKeys(writeList, curMap); err != nil {
		return err
	}
	// записываем результат
	r.repo.WriteList(writeList) // проводим операцию  с inmemory хранилищем
	r.journal.Write(opBytes)    // журналируем операцию
	return nil
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

func (r *RecordsInteractor) loadLogs(fList []string) error {
	for _, fName := range fList {
		ops, err := r.opr.loadFromFile(r.config.DirPath + fName)
		if err != nil {
			err = fmt.Errorf("Загрузка логов остановлена на файле `%s` с ошибкой `%s`", fName, err.Error())
			if r.config.AllowStartupErrLoadLogs {
				r.logger.Info(err)
				return nil
			}
			return err
		}
		if err := r.opr.doOperations(ops, r.repo); err != nil {
			return err
		}
	}
	return nil
}

func (r *RecordsInteractor) findLogsAfterCheckpoint(chpName string) ([]string, error) {
	logBarrier, err := strconv.ParseInt(strings.Replace(chpName, ".checkpoint", "", 1), 10, 64)
	if err != nil {
		return nil, err
	}
	logsNames, err := r.getFilesByExtList(".log")
	if err != nil {
		return nil, err
	}
	for i, logName := range logsNames {
		num, err := strconv.ParseInt(strings.Replace(logName, ".log", "", 1), 10, 64)
		if err != nil {
			return nil, err
		}
		if num > logBarrier {
			return logsNames[i : len(logsNames)-1], nil
		}
	}
	return make([]string, 0), nil
}

func (r *RecordsInteractor) findLatestCheckpoint() (string, error) {
	fNamesList, err := r.getFilesByExtList(".checkpoint")
	if err != nil {
		return "", err
	}
	if len(fNamesList) == 0 {
		return "", fmt.Errorf("Checkpoint not found (path: %s)", r.config.DirPath)
	}
	return fNamesList[len(fNamesList)-1], nil
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
