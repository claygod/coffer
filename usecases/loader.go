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
	riRepo domain.RecordsRepository
	fiRepo domain.RecordsRepository
}

func NewLoader(config *Config, lgr Logger, chp *checkpoint, opr *Operations, riRepo domain.RecordsRepository, fiRepo domain.RecordsRepository) *Loader {
	return &Loader{
		config: config,
		logger: lgr,
		chp:    chp,
		opr:    opr,
		riRepo: riRepo,
		fiRepo: fiRepo,
	}
}

func (l *Loader) LoadCheckpoint(chpName string) error {
	if err := l.chp.load(l.riRepo, chpName); err != nil { //загружаем последний checkpoint
		l.riRepo.Reset() //TODO:  это делается в checkpoins, но можно и продублировать (пока)
		return err
	} else if err := l.chp.load(l.fiRepo, chpName); err != nil {
		l.riRepo.Reset() //TODO:  это делается в checkpoins, но можно и продублировать (пока)
		l.fiRepo.Reset() //TODO:  это делается в checkpoins, но можно и продублировать (пока)
		return err
	}
	return nil
}

func (l *Loader) LoadLogs(fList []string) error {
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
		if err := l.opr.DoOperations(ops, l.riRepo); err != nil {
			return err
		}
		if err := l.opr.DoOperations(ops, l.fiRepo); err != nil {
			return err
		}
	}
	return nil
}
