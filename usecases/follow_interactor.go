package usecases

// Coffer
// Follow interactor
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	//"sync"

	"github.com/claygod/coffer/domain"
)

type FollowInteractor struct {
	//m               sync.Mutex
	logger          Logger
	config          *Config
	chp             *checkpoint
	opr             *Operations
	repo            domain.RecordsRepository
	filenamer       FileNamer
	changesCounter  int64
	lastFileNameLog string
	//lastFileNum     int64
	hasp Starter
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

) (*FollowInteractor, error) {
	fi := &FollowInteractor{
		logger:          logger,
		config:          config,
		chp:             chp,
		opr:             opr,
		repo:            repo,
		filenamer:       filenamer,
		lastFileNameLog: "-1.log", //TODO: in config
		hasp:            hasp,
	}
	// закачать последний чекпойнт и выставить его номер
	fChName, err := fi.filenamer.GetLatestFileName(extCheck + extPoint)
	if err != nil {
		return nil, err //TODO: тут надо реализовать кучу попыток с переходами к предыдущим номерам при неудаче!!!!!! - в отдельном методе
		//TODO: может быть битые чекпоинты переименовывать?
	} else if fChName != extCheck+extPoint && fChName != "" { //TODO: del `fChName != extCheck+extPoint`
		if err := fi.chp.load(fi.repo, fChName); err != nil { //загружаем последний checkpoint
			return nil, err
		}
		fi.lastFileNameLog = strings.Replace(fChName, extCheck+extPoint, extLog, 1)
	} /* else {
		fChName = "-1" + extCheck + extPoint
	}*/

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
	fmt.Println("F:запущен follow, list: ", list)
	for _, lFileName := range list {
		logFileName := f.config.DirPath + lFileName
		ops, err := f.opr.loadFromFile(logFileName) //тут ошибка возвращается только если нет каталога или ещё что-то подобное
		if err != nil {
			fmt.Println("F:err1: ", err)
			return err
		}
		if ops == nil { //значит файл пустой
			continue
		}
		if err := f.opr.DoOperations(ops, f.repo); err != nil {
			fmt.Println("F:err2: ", err)
			return err
		}
		fmt.Println("F:ops: ", len(ops), f.changesCounter, f.config.ChagesByCheckpoint, f.lastFileNameLog)
		f.addChangesCounter(ops)
		if f.changesCounter > f.config.ChagesByCheckpoint && logFileName != f.lastFileNameLog {
			fmt.Println("F:создал новый checkpoint: ", logFileName)
			if err := f.newCheckpoint(logFileName); err != nil {
				fmt.Println("F:что-то пошло не так: ", err)
				return err
			}
			if f.config.RemoveUnlessLogs {
				f.removingUselessLogs(logFileName)
			}
			f.changesCounter = 0
		}
		f.lastFileNameLog = logFileName
	}
	return nil
}

func (f *FollowInteractor) removingUselessLogs(lastLogPath string) { //TODO: учёт и удаление ненужных логов при усложнении вынести в отдельную сущность
	// f.m.Lock()
	// defer f.m.Unlock()
	fmt.Println("_______lastLogPath_______: ", lastLogPath)
	list1, err := f.filenamer.GetHalf(lastLogPath, false)
	if err != nil {
		f.logger.Warning(err)
	}
	//fmt.Println("_______list1_______: ", list1)
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
	//fmt.Println("_______list2_______: ", list2)
	for _, lgName := range list2 {
		err := os.Remove(f.config.DirPath + lgName) // на ошибки не смотрим, если какой-то файл случайно не удалится, не страшно
		if err != nil {
			f.logger.Warning(err)
		}
	}
}

func (f *FollowInteractor) addChangesCounter(ops []*domain.Operation) error {
	for _, op := range ops {
		f.changesCounter += int64(len(op.Body)) //считаем в байтах
	}
	return nil
}

func (f *FollowInteractor) findLatestLogs() ([]string, error) {
	//тут будем брать последние из filenamer
	fNamesList, err := f.filenamer.GetAfterLatest(f.lastFileNameLog)
	if err != nil {
		fmt.Println("F:err7:", err)
		return nil, err
	}
	ln := len(fNamesList)
	//fmt.Println("F:len(fNamesList):", len(fNamesList))
	if ln <= 1 { // последний лог мы тоже не берём чтобы не ткнуться в ещё наполняемый лог
		return make([]string, 0), nil
	}
	return fNamesList[0 : ln-2], nil

	// switch {
	// case ln == 0:
	// 	return fNamesList, nil
	// case ln == 1: // последний лог мы никогда не берём чтобы не ткнуться в ещё наполняемый лог
	// 	return make([]string, 0), nil
	// default:
	// 	// for num, fName := range fNamesList[:ln-1] { // если ничего не найдём, значит ещё не брали логи в работу
	// 	// 	if f.lastFileNameLog == fName {
	// 	// 		fNamesList = fNamesList[num : len(fNamesList)-num]
	// 	// 		break
	// 	// 	}
	// 	// }
	// }
	// //sort.Strings(fNamesList)
	// return fNamesList[], nil
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
