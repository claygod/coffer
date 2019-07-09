package coffer

// Coffer
// API tests
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"testing"
	"time"

	// "github.com/claygod/coffer/domain"
	// "github.com/claygod/coffer/services"
	// "github.com/claygod/coffer/services/filenamer"
	// "github.com/claygod/coffer/services/journal"
	// "github.com/claygod/coffer/services/repositories/handlers"
	// "github.com/claygod/coffer/services/repositories/records"
	"github.com/claygod/coffer/services/resources"
	// "github.com/claygod/coffer/services/startstop"
	"github.com/claygod/coffer/usecases"
	// "github.com/claygod/tools/logger"
	// "github.com/claygod/tools/porter"
)

func TestNewCoffer(t *testing.T) {
	ucCnf := &usecases.Config{
		FollowPause:             1 * time.Second,
		ChagesByCheckpoint:      100,
		DirPath:                 "test",
		AllowStartupErrLoadLogs: true,
		MaxKeyLength:            100,
		MaxValueLength:          10000,
	}
	rcCnf := &resources.Config{
		LimitMemory: 1000 * megabyte, // minimum available memory (bytes)
		LimitDisk:   1000 * megabyte, // minimum free disk space
		DirPath:     "test",
	}

	cnf := &Config{
		UsecasesConfig:      ucCnf,
		ResourcesConfig:     rcCnf,
		MaxRecsPerOperation: 100,
		//MaxKeyLength:        100,
		//MaxValueLength:      10000,
	}
	cof, err := New(cnf)
	if err != nil {
		t.Error(err)
	}
	if cof.Start() {
		defer cof.Stop()
	}
}
