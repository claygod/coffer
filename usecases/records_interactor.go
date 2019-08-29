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
	//"github.com/claygod/coffer/services/journal"
)

type RecordsInteractor struct {
	config     *Config
	logger     Logger
	loader     *Loader
	chp        *checkpoint
	opr        *Operations
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
	opr *Operations,
	coder *ReqCoder,
	repo domain.RecordsRepository,
	handlers domain.HandlersRepository,
	resControl Resourcer,
	porter Porter,
	journal Journaler,
	filenamer FileNamer,
	hasp Starter) (*RecordsInteractor, error) {

	r := &RecordsInteractor{
		config:     config,
		logger:     logger,
		loader:     loader,
		chp:        chp,
		opr:        opr,
		coder:      coder,
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
		return nil, err
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
		return nil, err
	}
	// выполнить все имеющиеся последующие логи
	if len(logsList) > 0 {
		// eсли последний по номеру не `checkpoint`, значит была аварийная остановка,
		// и нужно загрузить всё, что можно, сохранить, и только потом продолжить
		if err := r.loader.LoadLogs(logsList, r.repo); err != nil {
			return nil, err
		}
		if err := r.save(); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *RecordsInteractor) Start() bool {
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
	if err := r.save(); err != nil {
		r.logger.Error(err).
			Context("Object", "RecordsInteractor").
			Context("Method", "Stop").
			Send()
		//r.logger.Error(err)
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
	r.journal.Restart()
	return nil
}

func (r *RecordsInteractor) WriteList(req *ReqWriteList) (error, error) {
	if !r.hasp.Add() {
		return nil, fmt.Errorf("RecordsInteractor is stopped")
	}
	defer r.hasp.Done()
	// подготавливаем байтовую версию операции для лога
	opBytes, err := r.reqWriteListToLog(req)
	if err != nil {
		return nil, err
	}
	// проверяем, достаточно ли ресурсов (памяти, диска) для выполнения задачи
	if !r.resControl.GetPermission(int64(len(opBytes))) {
		return nil, fmt.Errorf("Insufficient resources (memory, disk)")
	}
	// блокируем нужные записи
	keys := r.getKeysFromMap(req.List)
	r.porter.Catch(keys)
	defer r.porter.Throw(keys)
	// выполняем
	r.repo.WriteList(req.List)                       // проводим операцию  с inmemory хранилищем
	if err := r.journal.Write(opBytes); err != nil { // журналируем операцию
		r.hasp.Stop()
		return err, nil
	}
	return nil, nil
}

func (r *RecordsInteractor) ReadList(req *ReqLoadList) (map[string][]byte, []string, error) {
	if !r.hasp.Add() {
		return nil, nil, fmt.Errorf("RecordsInteractor is stopped")
	}
	defer r.hasp.Done()
	// блокируем нужные записи
	r.porter.Catch(req.Keys)
	defer r.porter.Throw(req.Keys)
	// выполняем
	return r.repo.ReadList(req.Keys)
}

func (r *RecordsInteractor) DeleteList(req *ReqDeleteList) (error, error) {
	if !r.hasp.Add() {
		return fmt.Errorf("RecordsInteractor is stopped"), nil
	}
	defer r.hasp.Done()
	// подготавливаем байтовую версию операции для лога
	opBytes, err := r.reqDeleteListToLog(req)
	if err != nil {
		return nil, err
	}
	// проверяем, достаточно ли ресурсов (памяти, диска) для выполнения задачи
	if !r.resControl.GetPermission(int64(len(opBytes))) {
		return nil, fmt.Errorf("Insufficient resources (memory, disk)")
	}
	// блокируем нужные записи
	r.porter.Catch(req.Keys)
	defer r.porter.Throw(req.Keys)
	// выполняем
	wrn := r.repo.DelList(req.Keys)                  // при варнинге - не всё удалось удалить (каких-то ключей нет в базе)
	if err := r.journal.Write(opBytes); err != nil { // журналируем операцию
		r.hasp.Stop()
		return err, wrn
	}
	return nil, wrn //TODO: возвращать структуру с отчётом, что удалено а что нет
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

func (r *RecordsInteractor) Transaction(req *ReqTransaction) (error, error) { // interface{}, map[string][]byte, *domain.Handler
	if !r.hasp.Add() {
		return fmt.Errorf("RecordsInteractor is stopped"), nil
	}
	defer r.hasp.Done()
	// ищем хэндлер
	hdl, err := r.handlers.Get(req.HandlerName)
	if err != nil {
		return nil, err
	}
	handler := *hdl
	// подготавливаем байтовую версию операции для лога
	opBytes, err := r.reqTransactionToLog(req)
	if err != nil {
		return nil, err
	}
	// проверяем, достаточно ли ресурсов (памяти, диска) для выполнения задачи
	if !r.resControl.GetPermission(int64(len(opBytes))) {
		return nil, fmt.Errorf("Insufficient resources (memory, disk)")
	}
	// блокируем нужные записи
	r.porter.Catch(req.Keys)
	defer r.porter.Throw(req.Keys)
	// берём текущие значения в записях
	curMap, notFound, err := r.repo.ReadList(req.Keys)
	if err != nil {
		return nil, err
	} else if len(notFound) != 0 {
		return nil, fmt.Errorf("Records not found: %s", strings.Join(notFound, ", "))
	}
	// выполняем транзакцию
	writeList, err := handler(req.Value, curMap)
	if err != nil {
		return nil, err
	}
	// проверяем, нет ли "лишних" ключей в ответе
	if err := r.findExtraKeys(writeList, curMap); err != nil {
		return nil, err
	}
	// записываем результат
	r.repo.WriteList(writeList)                      // проводим операцию  с inmemory хранилищем
	if err := r.journal.Write(opBytes); err != nil { // журналируем операцию
		r.hasp.Stop()
		return err, nil
	}
	return nil, nil
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
