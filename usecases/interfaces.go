package usecases

// Coffer
// Interfaces
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"github.com/claygod/coffer/domain"
)

type Resourcer interface {
	GetPermission(int64) bool
}

type Porter interface {
	Catch([]string)
	Throw([]string)
}

type Logger interface {
	//Write(error)

	//Fatal(...interface{})
	Error(...interface{})
	Info(...interface{})
	Debug(...interface{})
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
}

type HandleStore interface {
	Get(string) (*domain.Handler, error)
	Set(string, *domain.Handler) error
}
