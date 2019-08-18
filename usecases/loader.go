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

func NewLoader(config *Config, lgr Logger, chp *checkpoint, opr *Operations) *Loader {
	return &Loader{
		config: config,
		logger: lgr,
		chp:    chp,
		opr:    opr,
	}
}

func (l *Loader) LoadLatestValidCheckpoint(chpList []string, repo domain.RecordsRepository) (string, error) {
	for i := len(chpList) - 1; i >= 0; i-- {
		fChName := l.config.DirPath + chpList[i]
		//fmt.Println("--------77--------currrrrrrrent ---- ", fChName)
		if fChName != extCheck+extPoint && fChName != "" { //TODO: del `fChName != extCheck+extPoint`
			if err := l.LoadCheckpoint(fChName, repo); err != nil { //загружаем последний checkpoint // chp.  load(r.repo, fChName)
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

func (l *Loader) LoadCheckpoint(chpName string, repo domain.RecordsRepository) error {
	if err := l.chp.load(repo, chpName); err != nil { //загружаем последний checkpoint
		repo.Reset() //TODO:  это делается в checkpoins, но можно и продублировать (пока)
		return err
	}
	return nil
}

func (l *Loader) LoadLogs(fList []string, repo domain.RecordsRepository) error {
	for _, fName := range fList {
		ops, err := l.opr.loadFromFile(l.config.DirPath + fName) //TODO: тут добавляем директорию к пути
		if err != nil {
			//err = fmt.Errorf("Загрузка логов остановлена на файле `%s` с ошибкой `%s`", fName, err.Error())
			if l.config.AllowStartupErrLoadLogs {
				l.logger.Info(err).
					Context("Object", "RecordsInteractor").
					Context("Method", "loadLogs").
					Send()
				return nil
			}
			return err
		}
		if err := l.opr.DoOperations(ops, repo); err != nil {
			return err
		}
	}
	return nil
}
