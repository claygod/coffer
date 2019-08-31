package usecases

// Coffer
// Checkpoint loader
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"

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

func (l *Loader) LoadLogs(fList []string, repo domain.RecordsRepository) error {
	counter := 0
	for _, fName := range fList {
		counter++
		ops, err, wrn := l.opr.loadFromFile(l.config.DirPath + fName) //TODO: тут добавляем директорию к пути
		if err != nil {
			return err
		}
		if wrn != nil {
			if len(fList) == counter && l.config.AllowStartupErrLoadLogs { //TODO: битый лог должен быть последним, если нет, то что-то не так
				l.logger.Info(wrn).
					Context("Object", "RecordsInteractor").
					Context("Method", "loadLogs").
					Send()
				return nil
			}
			return wrn
		}
		if err := l.opr.DoOperations(ops, repo); err != nil {
			return err
		}
	}
	return nil
}
