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

	"github.com/claygod/coffer/domain"
	//"github.com/claygod/coffer/services"

	// "github.com/claygod/coffer/services/filenamer"
	// "github.com/claygod/coffer/services/journal"
	// "github.com/claygod/coffer/services/repositories/handlers"
	// "github.com/claygod/coffer/services/repositories/records"
	//"github.com/claygod/coffer/services/journal"
	"github.com/claygod/coffer/services/resources"
	// "github.com/claygod/coffer/services/startstop"
	//"github.com/claygod/coffer/usecases"
	//"github.com/claygod/tools/logger"
	"github.com/sirupsen/logrus"
	// "github.com/claygod/tools/porter"
)

func TestNewOperations(t *testing.T) {
	ucCnf := &Config{
		FollowPause:             400 * time.Millisecond,
		LogsByCheckpoint:        2,
		DirPath:                 "../test/", // "/home/ed/goPath/src/github.com/claygod/coffer/test",
		AllowStartupErrLoadLogs: true,
		MaxKeyLength:            100,
		MaxValueLength:          10000,
	}
	rcCnf := &resources.Config{
		LimitMemory: 1000 * megabyte, // minimum available memory (bytes)
		LimitDisk:   1000 * megabyte, // minimum free disk space
		DirPath:     "../test/",      // "/home/ed/goPath/src/github.com/claygod/coffer/test"
	}
	resControl, err := resources.New(rcCnf)
	if err != nil {
		t.Error(err)
		return
	}
	hdl := newMockHandler()
	trn := NewTransaction(hdl)
	logger := logrus.New() //  logger.New(services.NewLog("Coffer "))
	reqCoder := NewReqCoder()

	//NewOperations(logger Logger, config *Config, reqCoder *ReqCoder, resControl Resourcer, trn *Transaction) *Operations
	NewOperations(logger, ucCnf, reqCoder, resControl, trn)
	//TODO oper.DoOperations()
}

type mockHandler struct {
}

func newMockHandler() *mockHandler {
	return &mockHandler{}
}

func (m *mockHandler) Get(handlerName string) (*domain.Handler, error) {
	hdl := domain.Handler(func(params interface{}, inMap map[string][]byte) (map[string][]byte, error) {
		return inMap, nil
	})
	return &hdl, nil //TODO
}
func (m *mockHandler) Set(handlerName string, handler *domain.Handler) {
	return //TODO
}
