package usecases

// Coffer
// Follow interactor
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/claygod/coffer/domain"
)

type FollowInteractor struct {
	//m               sync.Mutex
	logger Logger
	loader *Loader
	config *Config
	chp    *checkpoint
	//opr             *Operations
	repo            domain.RecordsRepository
	filenamer       FileNamer
	changesCounter  int64
	lastFileNameLog string
	hasp            Starter
}

func NewFollowInteractor(
	logger Logger,
	loader *Loader,
	config *Config,
	chp *checkpoint,
	//opr *Operations,
	repo domain.RecordsRepository,
	filenamer FileNamer,
	hasp Starter,

) (*FollowInteractor, error) {
	fi := &FollowInteractor{
		logger: logger,
		loader: loader,
		config: config,
		chp:    chp,
		//opr:             opr,
		repo:            repo,
		filenamer:       filenamer,
		lastFileNameLog: "-1.log", //TODO: in config
		hasp:            hasp,
	}

	chpList, err := fi.filenamer.GetHalf("-1"+extCheck+extPoint, true)
	if err != nil {
		return nil, err
	}
	fChName, err := fi.loader.LoadLatestValidCheckpoint(chpList, fi.repo) // загрузить последнюю валидную версию checkpoint
	if err != nil {
		fi.logger.Warning(err)
		fChName = "-1" + extCheck + extPoint
	}
	fi.lastFileNameLog = strings.Replace(fChName, extCheck+extPoint, extLog, 1) //  и выставить его номер

	return fi, nil
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
			f.logger.Error(err, "Method=worker", "Follow interactor is STOPPED!")
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
	if len(list) == 0 {
		return nil
	}
	//fmt.Println("F:запущен follow, list: ", list)
	err, wrn := f.loader.LoadLogs(list, f.repo)
	if err != nil {
		return err
	}
	if wrn != nil { // на битых файлах мы тоже стопаем
		return wrn
	}
	atomic.AddInt64(&f.changesCounter, int64(len(list))) // через атомик, чтобы при необходимости брать этот параметр, не было конкурентных проблем
	//f.changesCounter += len(list)
	logFileName := f.config.DirPath + list[len(list)-1]
	if atomic.LoadInt64((&f.changesCounter)) > f.config.LogsByCheckpoint && logFileName != f.lastFileNameLog {
		//fmt.Println("F:создал новый checkpoint: ", logFileName)
		if err := f.newCheckpoint(logFileName); err != nil {
			//fmt.Println("F:что-то пошло не так: ", err)
			return err
		}
		if f.config.RemoveUnlessLogs {
			f.removingUselessLogs(logFileName)
		}
		atomic.StoreInt64(&f.changesCounter, 0)
		//f.changesCounter = 0
	}
	f.lastFileNameLog = logFileName

	return nil
}

func (f *FollowInteractor) removingUselessLogs(lastLogPath string) { //TODO: учёт и удаление ненужных логов при усложнении вынести в отдельную сущность
	// f.m.Lock()
	// defer f.m.Unlock()
	list1, err := f.filenamer.GetHalf(lastLogPath, false)
	if err != nil {
		f.logger.Warning(err)
	}
	for _, lgName := range list1 {
		err := os.Remove(f.config.DirPath + lgName) // на ошибки не смотрим, если какой-то файл случайно не удалится, не страшно
		if err != nil {
			f.logger.Warning(err)
		}
	}

	list2, err := f.filenamer.GetHalf(strings.Replace(lastLogPath, extLog, extCheck+extPoint, 1), false)
	if err != nil {
		f.logger.Warning(err)
	}
	for _, lgName := range list2 {
		err := os.Remove(f.config.DirPath + lgName) // на ошибки не смотрим, если какой-то файл случайно не удалится, не страшно
		if err != nil {
			f.logger.Warning(err)
		}
	}
}

// func (f *FollowInteractor) addChangesCounter(ops []*domain.Operation) error {
// 	for _, op := range ops {
// 		f.changesCounter += int64(len(op.Body)) //считаем в байтах
// 	}
// 	return nil
// }

func (f *FollowInteractor) findLatestLogs() ([]string, error) {
	//тут будем брать последние из filenamer
	fNamesList, err := f.filenamer.GetHalf(f.lastFileNameLog, true)
	if err != nil {
		//fmt.Println("F:err7:", err)
		return nil, err
	}
	ln := len(fNamesList)
	//fmt.Println("F:len(fNamesList):", len(fNamesList))
	if ln <= 1 { // последний лог мы тоже не берём чтобы не ткнуться в ещё наполняемый лог
		return make([]string, 0), nil
	}
	return fNamesList[0 : ln-2], nil
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
