package usecases

// Coffer
// Operations tests
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	// "os"
	// "strconv"
	// "strings"
	"testing"
	"time"

	// "github.com/claygod/coffer/domain"
	"github.com/claygod/coffer/services"
	// "github.com/claygod/coffer/services/filenamer"
	// "github.com/claygod/coffer/services/journal"
	// "github.com/claygod/coffer/services/repositories/handlers"
	// "github.com/claygod/coffer/services/repositories/records"
	//"github.com/claygod/coffer/services/journal"
	"github.com/claygod/coffer/services/resources"
	// "github.com/claygod/coffer/services/startstop"
	//"github.com/claygod/coffer/usecases"
	"github.com/claygod/tools/logger"
	// "github.com/claygod/tools/porter"
)

func TestNewOperations(t *testing.T) {
	ucCnf := &Config{
		FollowPause:             400 * time.Millisecond,
		ChagesByCheckpoint:      2,
		DirPath:                 "./test/", // "/home/ed/goPath/src/github.com/claygod/coffer/test",
		AllowStartupErrLoadLogs: true,
		MaxKeyLength:            100,
		MaxValueLength:          10000,
	}
	rcCnf := &resources.Config{
		LimitMemory: 1000 * megabyte, // minimum available memory (bytes)
		LimitDisk:   1000 * megabyte, // minimum free disk space
		DirPath:     "./test/",       // "/home/ed/goPath/src/github.com/claygod/coffer/test"
	}
	resControl, err := resources.New(rcCnf)
	if err != nil {
		t.Error(err)
		return
	}
	trn := usecases.NewTransaction(c.handlers)
	logger := logger.New(services.NewLog("Coffer "))
	reqCoder := NewReqCoder()

	//NewOperations(logger Logger, config *Config, reqCoder *ReqCoder, resControl Resourcer, trn *Transaction) *Operations
	NewOperations(logger, ucCnf, reqCoder, resControl, trn)
}
