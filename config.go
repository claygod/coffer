package coffer

// Coffer
// Config
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"time"

	"github.com/claygod/coffer/services/journal"
	"github.com/claygod/coffer/services/resources"
	"github.com/claygod/coffer/usecases"
)

type Config struct {
	JournalConfig   *journal.Config
	UsecasesConfig  *usecases.Config
	ResourcesConfig *resources.Config

	//DataPath            string
	MaxRecsPerOperation int
	//MaxKeyLength        int
	//MaxValueLength      int
}

const (
	stateStopped int64 = iota
	stateStarted
	statePanic
)

const (
	logPrefix string = "Coffer "
	megabyte  int64  = 1024 * 1024
)

const (
	codeWriteList byte = iota //codeWrite
	codeTransaction
	codeDeleteList
)

const (
	defaultBatchSize              int   = 1000
	defaultLimitRecordsPerLogfile int64 = 1000

	defaultFollowPause             time.Duration = 60 * time.Second
	defaultLogsByCheckpoint        int64         = 10
	defaultAllowStartupErrLoadLogs bool          = true
	defaultMaxKeyLength            int           = 100
	defaultMaxValueLength          int           = 10000
	defaultRemoveUnlessLogs        bool          = true

	defaultLimitMemory int64 = 100 * megabyte
	defaultLimitDisk   int64 = 1000 * megabyte

	defaultMaxRecsPerOperation int = 1000
)
