package usecases

// Coffer
// Config
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"time"
)

type Config struct {
	FollowPause             time.Duration
	ChagesByCheckpoint      int64
	DirPath                 string
	AllowStartupErrLoadLogs bool
	MaxKeyLength            int //  = int(uint64(1)<<16) - 1
	MaxValueLength          int //  = int(uint64(1)<<48) - 1
}

const (
	stateStopped int64 = iota
	stateStarted
	statePanic
)

const (
	megabyte int64 = 1024 * 1024
)

const (
	codeWriteList byte = iota //codeWrite
	codeTransaction
	codeDeleteList
)
