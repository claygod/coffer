package coffer

// Coffer
// API
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	//"time"

	"github.com/claygod/coffer/domain"
	//"github.com/claygod/coffer/services"
	"github.com/claygod/coffer/services/filenamer"
	"github.com/claygod/coffer/services/journal"

	//"github.com/claygod/coffer/services/logger"
	"github.com/claygod/coffer/services/porter"
	//"github.com/claygod/coffer/services/repositories/handlers"
	"github.com/claygod/coffer/services/repositories/records"
	"github.com/claygod/coffer/services/resources"
	"github.com/claygod/coffer/services/startstop"
	"github.com/claygod/coffer/usecases"

	"github.com/sirupsen/logrus"
)

/*
Coffer - Simple ACID* key-value database.
*/
type Coffer struct {
	config        *Config
	logger        usecases.Logger
	porter        usecases.Porter
	resControl    *resources.ResourcesControl
	handlers      domain.HandlersRepository
	recInteractor *usecases.RecordsInteractor
	folInteractor *usecases.FollowInteractor
	panicRecover  func()
	hasp          usecases.Starter
}

func new(config *Config, hdls domain.HandlersRepository) (*Coffer, error, error) {
	//TODO: check received config
	resControl, err := resources.New(config.ResourcesConfig)
	if err != nil {
		return nil, err, nil
	}
	//TODO: not remove this code!!
	// if hdls == nil {
	// 	hdls = handlers.New()
	// }
	logger := logrus.New()

	c := &Coffer{
		config:     config,
		logger:     logger.WithField("Object", "Coffer"),
		porter:     porter.New(),
		resControl: resControl,
		handlers:   hdls,
		hasp:       startstop.New(),
	}

	c.panicRecover = func() {
		if r := recover(); r != nil {
			c.logger.Error(r)
		}
	}

	alarmFunc := func(err error) { // для журнала
		logger.WithField("Object", "Journal").WithField("Method", "Write").Error(err)
	}
	riRepo := records.New()
	fiRepo := records.New()
	reqCoder := usecases.NewReqCoder()
	fileNamer := filenamer.NewFileNamer(c.config.UsecasesConfig.DirPath)
	trn := usecases.NewTransaction(c.handlers)
	chp := usecases.NewCheckpoint(c.config.UsecasesConfig)
	//opr := usecases.NewOperations(c.logger, c.config.UsecasesConfig, reqCoder, resControl, trn)
	ldr := usecases.NewLoader(config.UsecasesConfig, logger.WithField("Object", "Loader"), chp, reqCoder, resControl, trn)
	jrn, err := journal.New(c.config.JournalConfig, fileNamer, alarmFunc)
	if err != nil {
		return nil, err, nil
	}
	ri, err, wrn := usecases.NewRecordsInteractor(
		c.config.UsecasesConfig,
		logger.WithField("Object", "RecordsInteractor"), //c.logger,
		ldr,
		chp,
		//opr,
		trn,
		reqCoder,
		riRepo, //recordsRepo,
		c.handlers,
		resControl,
		//c.porter,
		jrn,
		fileNamer,
		startstop.New(),
	)
	if err != nil {
		return nil, err, wrn
	}
	c.recInteractor = ri

	fi, err := usecases.NewFollowInteractor(
		logger.WithField("Object", "FollowInteractor"),
		ldr,
		c.config.UsecasesConfig,
		chp, //*checkpoint,
		//opr,                     // *operations,
		fiRepo, //recordsRepo,
		fileNamer,
		startstop.New(),
	)
	if err != nil {
		return nil, err, nil
	}
	c.folInteractor = fi

	return c, nil, nil
}

/*
Start - database launch
*/
func (c *Coffer) Start() bool {
	//TODO: при аварийной остановке нужно ли иметь возможность запускаться вновь?
	// (StopForever, Concrete - в старт-стоп добавить) возможно, правильный выход - пересоздание и запуск
	defer c.panicRecover()

	c.resControl.Start()
	c.recInteractor.Start()
	c.folInteractor.Start()
	c.hasp.Start()

	//TODO: not remove this code!!
	// if !c.resControl.Start() {
	// 	return false
	// }
	// if !c.recInteractor.Start() {
	// 	c.resControl.Stop()
	// 	return false
	// }
	// //fmt.Println("recInteractor.Start")
	// if !c.folInteractor.Start() {
	// 	c.resControl.Stop()
	// 	c.recInteractor.Stop()
	// 	return false
	// }
	// if !c.hasp.Start() {
	// 	c.resControl.Stop()
	// 	c.recInteractor.Stop()
	// 	c.folInteractor.Stop()
	// 	return false
	// }

	//fmt.Println("folInteractor.Start")
	return true
}

/*
Stop - database stop
*/
func (c *Coffer) Stop() bool {
	if c.hasp.IsReady() {
		return true // already stopped
	}

	defer c.panicRecover()
	//TODO: not remove this code!!
	// if !c.hasp.Block() {
	// 	return false
	// }
	c.hasp.Block()
	defer c.hasp.Unblock()
	c.resControl.Stop()
	c.folInteractor.Stop()
	c.recInteractor.Stop()
	//TODO: not remove this code!!
	// if !c.resControl.Stop() {
	// 	return false
	// }
	// if !c.folInteractor.Stop() {
	// 	c.resControl.Start()
	// 	return false
	// }
	// if !c.recInteractor.Stop() {
	// 	c.resControl.Start()
	// 	c.folInteractor.Start()
	// 	return false
	// }
	return true
}

// /*
// StopHard - immediate stop of the database, without waiting for the stop of internal processes.
// The operation is quick, but extremely dangerous.
// */
// func (c *Coffer) StopHard() error {
// 	defer c.panicRecover()
// 	var errOut error
// 	c.hasp.Block()
// 	c.folInteractor.Stop()
// 	c.recInteractor.Stop()
// 	//TODO: not remove this code!!
// 	// if !c.hasp.Block() {
// 	// 	errOut = fmt.Errorf("Hasp is not stopped.")
// 	// }
// 	// if !c.folInteractor.Stop() {
// 	// 	errOut = fmt.Errorf("%v Follow Interactor is not stopped.", errOut)
// 	// }
// 	// if !c.recInteractor.Stop() {
// 	// 	errOut = fmt.Errorf("%v Records Interactor is not stopped.", errOut)
// 	// }
// 	return errOut
// }
