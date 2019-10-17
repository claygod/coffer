package usecases

// Coffer
// Config
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"time"
)

type Config struct {
	FollowPause             time.Duration
	LogsByCheckpoint        int64
	DirPath                 string
	AllowStartupErrLoadLogs bool
	RemoveUnlessLogs        bool // deleting logs after they hit the checkpoint
	MaxKeyLength            int  //  = int(uint64(1)<<16) - 1
	MaxValueLength          int  //  = int(uint64(1)<<48) - 1
}

const (
	stateStopped int64 = iota
	stateStarted
	statePanic
)

const (
	extLog   string = ".log"
	extCheck string = ".check"
	extPoint string = "point"
	megabyte int64  = 1024 * 1024
)

const (
	codeWriteList byte = iota //codeWrite
	codeTransaction
	codeDeleteListStrict
	codeDeleteListOptional
)
