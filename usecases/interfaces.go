package usecases

// Coffer
// Interfaces
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"github.com/claygod/coffer/domain"
	"github.com/claygod/tools/logger"
)

type Resourcer interface {
	GetPermission(int64) bool
}

type Porter interface {
	Catch([]string)
	Throw([]string)
}

type Logger interface {
	//Fatal(...interface{})
	Error(interface{}) *logger.Logger
	Warning(interface{}) *logger.Logger
	Info(interface{}) *logger.Logger
	Context(string, interface{}) *logger.Logger
	Send() (int, error)
	//Debug(...interface{})
}

type Journaler interface {
	Write([]byte)
	Close()
}

type Starter interface {
	Start() bool
	Stop() bool
	Add() bool
	Done() bool
	Total() int64
	IsReady() bool
	Block() bool
	Unblock() bool
}

type HandleStore interface {
	Get(string) (*domain.Handler, error)
	Set(string, *domain.Handler) error
}

type FileNamer interface {
	GetNewFileName(ext string) (string, error)
	GetAfterLatest(last string) ([]string, error)
	GetLatestFileName(ext string) (string, error)
}
