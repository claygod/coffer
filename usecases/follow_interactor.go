package usecases

// Coffer
// Follow interactor
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"os"
	"time"

	"github.com/claygod/coffer/domain"
)

type FollowInteractor struct {
	logger             Logger
	chp                *checkpoint
	repo               domain.RecordsRepository
	pause              time.Duration
	chagesByCheckpoint int64
	lastLogNum         int64
	changesCounter     int64
	hasp               Starter
}

func (f *FollowInteractor) Start() bool {
	if !f.Start() {
		return false
	}
	go f.worker()
	return true
}

func (f *FollowInteractor) Stop() bool {
	if !f.Stop() {
		return false
	}
	return true
}

func (f *FollowInteractor) worker() {
	for {
		if f.hasp.IsReady() {
			return
		}
		f.hasp.Add()
		f.follow()
		f.hasp.Done()
		time.Sleep(f.pause)
	}
}

func (f *FollowInteractor) follow() {
	// logsNamesList := f.findLatestLogs()
	for _, logFileName := range f.findLatestLogs() {
		logFile, err := os.Open(logFileName)
		if err != nil {
			f.logger.Write(err)
			return
		}
		ops, err := f.loadOperationsFromFile(logFile)
		logFile.Close()
		if err != nil {
			f.logger.Write(err)
			return
		}
		if err := f.doOperations(ops); err != nil {
			f.logger.Write(err)
			return
		}
		f.setLastNum(logFileName)
		if f.changesCounter > f.chagesByCheckpoint {
			if err := f.newCheckpoint(); err != nil {
				f.logger.Write(err)
				return
			}
			f.changesCounter = 0
		}
	}
	return
}

func (f *FollowInteractor) doOperations(ops []*domain.Operation) error {
	//TODO:
	//for  f.changesCounter++
	return nil
}

func (f *FollowInteractor) loadOperationsFromFile(file *os.File) ([]*domain.Operation, error) {
	//TODO:
	return nil, nil
}

func (f *FollowInteractor) findLatestLogs() []string {
	//TODO:
	return nil
}

func (f *FollowInteractor) getLogsList() []string {
	//TODO:
	return nil
}

func (f *FollowInteractor) setLastNum(logFileName string) error {
	//TODO:
	return nil
}

func (f *FollowInteractor) newCheckpoint() error {
	//TODO:
	return nil
}
