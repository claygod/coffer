package usecases

// Coffer
// Operations tests
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"testing"
	"time"

	"github.com/claygod/coffer/domain"
	"github.com/claygod/coffer/services/resources"
	//"github.com/sirupsen/logrus"
)

func TestNewOperations(t *testing.T) {
	ucCnf := &Config{
		FollowPause:             400 * time.Millisecond,
		LogsByCheckpoint:        2,
		DirPath:                 "../test/",
		AllowStartupErrLoadLogs: true,
		MaxKeyLength:            100,
		MaxValueLength:          10000,
	}
	rcCnf := &resources.Config{
		LimitMemory: 1000 * megabyte, // minimum available memory (bytes)
		LimitDisk:   1000 * megabyte, // minimum free disk space
		DirPath:     "../test/",
	}
	resControl, err := resources.New(rcCnf)
	if err != nil {
		t.Error(err)
		return
	}
	hdl := newMockHandler()
	trn := NewTransaction(hdl)
	//logger := logrus.New() //  logger.New(services.NewLog("Coffer "))
	reqCoder := NewReqCoder()
	NewOperations(ucCnf, reqCoder, resControl, trn)
}

type mockHandler struct {
}

func newMockHandler() *mockHandler {
	return &mockHandler{}
}

func (m *mockHandler) Get(handlerName string) (*domain.Handler, error) {
	hdl := domain.Handler(func(params []byte, inMap map[string][]byte) (map[string][]byte, error) {
		return inMap, nil
	})
	return &hdl, nil //TODO
}
func (m *mockHandler) Set(handlerName string, handler *domain.Handler) {
	return //TODO
}
