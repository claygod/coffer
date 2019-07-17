package usecases

// Coffer
// Follow interactor
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"time"

	"github.com/claygod/coffer/domain"
)

type FollowInteractor struct {
	logger          Logger
	config          *Config
	chp             *checkpoint
	opr             *Operations
	repo            domain.RecordsRepository
	filenamer       FileNamer
	changesCounter  int64
	lastFileNameLog string
	lastFileNum     int64
	hasp            Starter
}

func NewFollowInteractor(
	logger Logger,
	config *Config,
	chp *checkpoint,
	opr *Operations,
	repo domain.RecordsRepository,
	filenamer FileNamer,
	//changesCounter  int64,
	//lastFileNameLog string,
	hasp Starter,

) *FollowInteractor {
	fi := &FollowInteractor{
		logger:    logger,
		config:    config,
		chp:       chp,
		opr:       opr,
		repo:      repo,
		filenamer: filenamer,
		hasp:      hasp,
	}
	//TODO: закачать последний чекпойнт и выставить его номер
	return fi
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
			f.logger.Error(err).
				Context("Object", "FollowInteractor").
				Context("Method", "worker").
				Context("Message", "Follow interactor is STOPPED!").
				Send()
			// f.logger.Error(err)
			// f.logger.Error(fmt.Errorf("Follow interactor is STOPPED!"))
			return
		}
		f.hasp.Done()
		time.Sleep(f.config.FollowPause)
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
		if err := f.opr.DoOperations(ops, f.repo); err != nil {
			return err
		}
		f.addChangesCounter(ops)
		if f.changesCounter > f.config.ChagesByCheckpoint && logFileName != f.lastFileNameLog {
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

func (f *FollowInteractor) findLatestLogs() ([]string, error) {
	fNamesList, err := f.getFilesByExtList(extLog)
	if err != nil {
		return nil, err
	}
	ln := len(fNamesList)
	switch {
	case ln == 0:
		return fNamesList, nil
	case ln == 1: // последний лог мы никогда не берём чтобы не ткнуться в ещё наполняемый лог
		return make([]string, 0), nil
	default:
		for num, fName := range fNamesList[:ln-1] { // если ничего не найдём, значит ещё не брали логи в работу
			if f.lastFileNameLog == fName {
				fNamesList = fNamesList[num : len(fNamesList)-num]
				break
			}
		}
	}
	sort.Strings(fNamesList)
	return fNamesList, nil
}

func (f *FollowInteractor) getFilesByExtList(ext string) ([]string, error) {
	files, err := ioutil.ReadDir(f.config.DirPath)
	if err != nil {
		return nil, err
	}
	list := make([]string, 0, len(files))
	for _, fl := range files {
		if strings.HasSuffix(fl.Name(), ext) {
			list = append(list, f.config.DirPath+fl.Name())
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
	return strings.Replace(logFileName, extLog, extCheck, 1)
	// strNum := strconv.FormatInt(f.lastNumCheckpoint, 10)
	// strNum += ".check"
	// f.lastNumCheckpoint++
	// return strNum
}
