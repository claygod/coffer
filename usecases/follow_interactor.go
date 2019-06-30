package usecases

// Coffer
// Follow interactor
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"bytes"
	//"encoding/gob"
	"fmt"

	//"io"
	"io/ioutil"
	//"os"
	"strings"
	"time"

	"github.com/claygod/coffer/domain"
)

type FollowInteractor struct {
	logger Logger
	chp    *checkpoint
	opr    *operations
	//trn                *transaction
	//reqCoder *ReqCoder
	repo domain.RecordsRepository
	//handlers HandleStore
	//resoureControl     Resourcer
	pause              time.Duration //TODO: в отдельный конфиг
	chagesByCheckpoint int64         //TODO: в отдельный конфиг - сколько изменений допустимо в одном чекпойнте
	changesCounter     int64
	lastFileNameLog    string
	logDir             string //TODO: в отдельный конфиг

	hasp Starter
}

func (f *FollowInteractor) Start() bool {
	if !f.hasp.Start() {
		return false
	}
	go f.worker()
	return true
}

func (f *FollowInteractor) Stop() bool {
	if !f.hasp.Stop() {
		return false
	}
	return true
}

/*
worker - циклическое приближение чекпойнтов к актуальному состоянию.
При любой ошибке работа останавливается (как минимум до перезагрузки).
*/
func (f *FollowInteractor) worker() {
	for {
		if f.hasp.IsReady() {
			return
		}
		f.hasp.Add()
		if err := f.follow(); err != nil {
			f.hasp.Done()
			f.Stop()
			f.hasp.Block()
			f.logger.Error(err)
			f.logger.Error(fmt.Errorf("Follow interactor is STOPPED!"))
			return
		}
		f.hasp.Done()
		time.Sleep(f.pause)
	}
}

func (f *FollowInteractor) follow() error {
	list, err := f.findLatestLogs()
	if err != nil {
		return err
	}
	for _, logFileName := range list {
		ops, err := f.opr.loadFromFile(logFileName)
		if err != nil {
			return err
		}
		if err := f.opr.doOperations(ops, f.repo); err != nil {
			return err
		}
		f.addChangesCounter(ops)
		if f.changesCounter > f.chagesByCheckpoint && logFileName != f.lastFileNameLog {
			if err := f.newCheckpoint(logFileName); err != nil {
				return err
			}
			f.changesCounter = 0
			f.lastFileNameLog = logFileName
		}
	}
	return nil
}

func (f *FollowInteractor) addChangesCounter(ops []*domain.Operation) error {
	for _, op := range ops {
		f.changesCounter += int64(len(op.Body)) //считаем в байтах
	}
	return nil
}

// func (f *FollowInteractor) doOperations(ops []*domain.Operation) error {
// 	for _, op := range ops {
// 		if !f.resoureControl.GetPermission(int64(len(op.Body))) {
// 			return fmt.Errorf("Operation code %d, len(body)=%d, Not permission!", op.Code, len(op.Body))
// 		}
// 		switch op.Code {
// 		case codeWriteList:
// 			reqWL, err := f.reqCoder.ReqWriteListDecode(op.Body)
// 			if err != nil {
// 				return err
// 			} else if err := f.repo.SetRecords(f.convertReqWriteListToRecords(reqWL)); err != nil {
// 				return err
// 			}
// 		case codeTransaction:
// 			reqTr, err := f.reqCoder.ReqTransactionDecode(op.Body)
// 			if err != nil {
// 				return err
// 			}
// 			if err := f.trn.doOperationTransaction(reqTr, f.repo); err != nil {
// 				return err
// 			}
// 		case codeDeleteList:
// 			reqDL, err := f.reqCoder.ReqDeleteListDecode(op.Body)
// 			if err != nil {
// 				return err
// 			} else if err := f.repo.DelRecords(reqDL.Keys); err != nil {
// 				return err
// 			}
// 		default:
// 			return fmt.Errorf("Unknown operation `%d`", op.Code)
// 		}
// 		f.changesCounter += int64(len(op.Body)) //считаем в байтах
// 	}
// 	return nil
// }

func (f *FollowInteractor) convertReqWriteListToRecords(req *ReqWriteList) []*domain.Record {
	recs := make([]*domain.Record, 0, len(req.List))
	for key, value := range req.List {
		rec := &domain.Record{
			Key:   key,
			Value: value,
		}
		recs = append(recs, rec)
	}
	return recs
}

func (f *FollowInteractor) findLatestLogs() ([]string, error) {
	fNamesList, err := f.getFilesByExtList(".log")
	if err != nil {
		return nil, err
	}
	ln := len(fNamesList)
	switch { // последний лог мы никогда не берём чтобы не ткнуться в ещё наполняемый лог
	case ln == 0:
		return fNamesList, nil
	case ln == 1:
		return make([]string, 0), nil
	default:
		for num, fName := range fNamesList[:ln-1] { // если ничего не найдём, значит ещё не брали логи в работу
			if f.lastFileNameLog == fName {
				fNamesList = fNamesList[num : len(fNamesList)-num]
				break
			}
		}
	}

	return fNamesList, nil
}

func (f *FollowInteractor) getFilesByExtList(ext string) ([]string, error) {
	files, err := ioutil.ReadDir(f.logDir)
	if err != nil {
		return nil, err
	}
	list := make([]string, 0, len(files))
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ext) {
			list = append(list, f.Name())
		}
	}
	return list, nil
}

func (f *FollowInteractor) newCheckpoint(logFileName string) error {
	if err := f.chp.save(f.repo, f.getNewCheckpointName(logFileName)); err != nil {
		return err
	}
	return nil
}

func (f *FollowInteractor) getNewCheckpointName(logFileName string) string { // просто меняем расширение файла
	// strs := strings.Split(logFileName, ".")
	// return strs[0] + ".check"

	return strings.Replace(logFileName, ".log", ".check", 1)

	// strNum := strconv.FormatInt(f.lastNumCheckpoint, 10)
	// strNum += ".check"
	// f.lastNumCheckpoint++
	// return strNum
}
