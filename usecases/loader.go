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
	return &Loader{
		config: config,
		logger: lgr,
		chp:    chp,
		opr:    NewOperations(config, reqCoder, resControl, trn),
	}
}

func (l *Loader) LoadLatestValidCheckpoint(chpList []string, repo domain.RecordsRepository) (string, error) {
	for i := len(chpList) - 1; i >= 0; i-- {
		fChName := chpList[i]
		if fChName != extCheck+extPoint && fChName != "" {
			if err := l.loadCheckpoint(fChName, repo); err != nil { // load the last checkpoint
				l.logger.Info(err)
			} else {
				return fChName, nil
			}
		}
	}
	return "-1" + extCheck + extPoint, nil
}

func (l *Loader) loadCheckpoint(chpName string, repo domain.RecordsRepository) error {
	if err := l.chp.load(repo, l.config.DirPath+chpName); err != nil { // load the last checkpoint
		repo.Reset() //TODO:  this is done in checkpoints, but you can duplicate (for now)
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
		ops, err, wrn := l.opr.loadFromFile(l.config.DirPath + fName)
		if err != nil {
			return err, wrn
		}
		if wrn != nil {
			wr = wrn
			switch counter { // two options, as sometimes there will be a log with zero content last
			case len(fList):
				if !l.config.AllowStartupErrLoadLogs {
					return fmt.Errorf("The spoiled log. l.config.AllowStartupErrLoadLogs == false"), wrn
				} else {
					brk = true
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
					brk = true
				}
			default:
				return fmt.Errorf("The spoiled log (%s) .", l.config.DirPath+fName), wrn
			}
		}
		if err := l.opr.DoOperations(ops, repo); err != nil {
			return err, wrn
		}
		if brk {
			break
		}
	}
	return nil, wr
}
