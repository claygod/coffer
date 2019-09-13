package usecases

// Coffer
// Checkpoint loader
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"os"

	"github.com/claygod/coffer/domain"
)

type Loader struct {
	config *Config
	logger Logger
	chp    *checkpoint
	opr    *Operations
}

func NewLoader(config *Config, lgr Logger, chp *checkpoint, reqCoder *ReqCoder, resControl Resourcer, trn *Transaction) *Loader {
	//oper := NewOperations(lgr, config, reqCoder, resControl, trn)

	return &Loader{
		config: config,
		logger: lgr,
		chp:    chp,
		opr:    NewOperations(lgr, config, reqCoder, resControl, trn),
	}
}

func (l *Loader) LoadLatestValidCheckpoint(chpList []string, repo domain.RecordsRepository) (string, error) {
	for i := len(chpList) - 1; i >= 0; i-- {
		fChName := chpList[i] // l.config.DirPath +
		//fmt.Println("--------77--------currrrrrrrent ---- ", fChName)
		if fChName != extCheck+extPoint && fChName != "" { //TODO: del `fChName != extCheck+extPoint`
			if err := l.loadCheckpoint(fChName, repo); err != nil { //загружаем последний checkpoint // chp.  load(r.repo, fChName)
				l.logger.Info(err)
			} else {
				//fmt.Println("-------77---------Найден чекпоинт ", fChName)
				//fmt.Println("--------77--------Вот сколько теперь записей: ", repo.CountRecords())
				return fChName, nil
			}
		}
	}
	return "-1" + extCheck + extPoint, nil
}

func (l *Loader) loadCheckpoint(chpName string, repo domain.RecordsRepository) error {
	if err := l.chp.load(repo, l.config.DirPath+chpName); err != nil { //загружаем последний checkpoint
		repo.Reset() //TODO:  это делается в checkpoins, но можно и продублировать (пока)
		return err
	}
	return nil
}

func (l *Loader) LoadLogs(fList []string, repo domain.RecordsRepository) (error, error) {
	counter := 0
	var wr error
	for _, fName := range fList {
		brk := false
		counter++
		ops, err, wrn := l.opr.loadFromFile(l.config.DirPath + fName) //TODO: тут добавляем директорию к пути
		if err != nil {
			return err, wrn
		}
		if wrn != nil {
			wr = wrn
			//fmt.Println(len(fList), counter, fList)
			switch counter { // два варианта, т.к. иногда будет лог с нулевым содержимым последним
			case len(fList):
				if !l.config.AllowStartupErrLoadLogs {
					return fmt.Errorf("The spoiled log. l.config.AllowStartupErrLoadLogs == false"), wrn
				} else {
					brk = true // return nil, wrn
				}
			case len(fList) - 1:
				stat, err := os.Stat(l.config.DirPath + fList[len(fList)-1])
				if err != nil {
					return err, wrn
				}
				if stat.Size() != 0 {
					return fmt.Errorf("The spoiled log (%s) is not the last, after it there is one more log file.",
						l.config.DirPath+fName), wrn
				}
				if !l.config.AllowStartupErrLoadLogs {
					return fmt.Errorf("The spoiled log. l.config.AllowStartupErrLoadLogs == false"), wrn
				} else {
					brk = true //return nil, wrn
				}
			default:
				return fmt.Errorf("The spoiled log (%s) .", l.config.DirPath+fName), wrn
			}
		}
		// if len(fList) != counter || !l.config.AllowStartupErrLoadLogs {
		// 	return fmt.Errorf("The spoiled log (%s) is not the last, after it there is one more log file. OR l.config.AllowStartupErrLoadLogs == false",
		// 		l.config.DirPath+fName), wrn

		// if wrn != nil {
		// 	if len(fList) == counter && l.config.AllowStartupErrLoadLogs { //TODO: битый лог должен быть последним, если нет, то что-то не так
		// 		l.logger.Info(wrn).
		// 			Context("Object", "RecordsInteractor").
		// 			Context("Method", "loadLogs").
		// 			Send()
		// 		return nil, wrn
		// 	}
		// 	return wrn
		// }
		if err := l.opr.DoOperations(ops, repo); err != nil {
			return err, wrn
		}
		if brk {
			break
		}
	}
	return nil, wr
}
