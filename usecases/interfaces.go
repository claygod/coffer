package usecases

// Coffer
// Interfaces
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"github.com/claygod/coffer/domain"
	//"github.com/claygod/coffer/services/logger"
)

/*
Resourcer - interface for indicator of the status of the physical memory (and disk) of the device.
*/
type Resourcer interface {
	GetPermission(int64) bool
}

/*
Porter - interface for regulates access to resources by keys.
*/
type Porter interface {
	Catch([]string)
	Throw([]string)
}

/*
Logger - interface for logs.
*/
type Logger interface {
	//Fatal(...interface{})
	Error(...interface{})   //*logger.Logger
	Warning(...interface{}) // *logger.Logger
	Info(...interface{})    //*logger.Logger
	//Context(string, interface{})// *logger.Logger
	//Send() (int, error)
	//Debug(...interface{})
}

/*
Journaler - interface for journal.
*/
type Journaler interface {
	Write([]byte) error
	Start() error
	Stop()
	//Close()
	Restart()
}

/*
Starter - interface for StartStop.
*/
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

/*
HandleStore - interface for handlers store.
*/
type HandleStore interface {
	Get(string) (*domain.Handler, error)
	Set(string, *domain.Handler)
}

/*
FileNamer - interface for logs names creator.
*/
type FileNamer interface {
	GetNewFileName(ext string) (string, error)
	GetAfterLatest(last string) ([]string, error)
	GetHalf(last string, more bool) ([]string, error)
	GetLatestFileName(ext string) (string, error)
}
