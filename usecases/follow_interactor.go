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

/*
FollowInteractor - after the database is launched, it writes all operations to the log. As a result,
the log can grow very much. If in the end, at the end of the application, the database is correctly stopped,
a new checkpoint will appear, and at the next start, the data will be taken from it.
However, the stop may not be correct, and a new checkpoint will not be created.

In this case, at a new start, the database will be forced to load the old checkpoint, and re-perform all operations
that were completed and recorded in the log. This can turn out to be quite significant in time, and as a result,
the database will take longer to load, which is not always acceptable for applications.

That is why there is a follower mechanism in the database that methodically goes through the logs in the process of
the database and periodically creates checkpoints that are much closer to the current moment.
Also, the follower has the function of cleaning old logs and checkpoints to free up space on your hard drive.
*/
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

/*
NewFollowInteractor - create new interactor.
*/
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

/*
Start - launch FollowInteractor.
*/
func (f *FollowInteractor) Start() bool {
	if !f.hasp.Start() {
		return false
	}
	go f.worker()
	return true
}

/*
Stop - stop FollowInteractor.
*/
func (f *FollowInteractor) Stop() bool {
	if !f.hasp.Stop() {
		return false
	}
	return true
}

/*
worker - cyclic approximation of checkpoints to the current state.
If any error occurs, operation stops (at least until a reboot).
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
	err, wrn := f.loader.LoadLogs(list, f.repo)
	if err != nil {
		return err
	}
	if wrn != nil { // we also stop on broken files
		return wrn
	}
	atomic.AddInt64(&f.changesCounter, int64(len(list)))
	logFileName := f.config.DirPath + list[len(list)-1]
	if atomic.LoadInt64((&f.changesCounter)) > f.config.LogsByCheckpoint && logFileName != f.lastFileNameLog {
		if err := f.newCheckpoint(logFileName); err != nil {
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

func (f *FollowInteractor) removingUselessLogs(lastLogPath string) {
	// f.m.Lock()
	// defer f.m.Unlock()
	list1, err := f.filenamer.GetHalf(lastLogPath, false)
	if err != nil {
		f.logger.Warning(err)
	}
	for _, lgName := range list1 {
		err := os.Remove(f.config.DirPath + lgName) // we don’t look at errors if some file is not deleted accidentally, it’s not scary
		if err != nil {
			f.logger.Warning(err)
		}
	}

	list2, err := f.filenamer.GetHalf(strings.Replace(lastLogPath, extLog, extCheck+extPoint, 1), false)
	if err != nil {
		f.logger.Warning(err)
	}
	for _, lgName := range list2 {
		err := os.Remove(f.config.DirPath + lgName) // we don’t look at errors if some file is not deleted accidentally, it’s not scary
		if err != nil {
			f.logger.Warning(err)
		}
	}
}

func (f *FollowInteractor) findLatestLogs() ([]string, error) {
	fNamesList, err := f.filenamer.GetHalf(f.lastFileNameLog, true)
	if err != nil {
		return nil, err
	}
	ln := len(fNamesList)
	if ln <= 1 { // we don’t take the last log so as not to stumble into the log that is still being filled
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

func (f *FollowInteractor) getNewCheckpointName(logFileName string) string {
	// strs := strings.Split(logFileName, ".")
	// return strs[0] + ".check"
	return strings.Replace(logFileName, extLog, extCheck, 1)
	// strNum := strconv.FormatInt(f.lastNumCheckpoint, 10)
	// strNum += ".check"
	// f.lastNumCheckpoint++
	// return strNum
}
