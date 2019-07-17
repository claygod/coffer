package coffer

// Coffer
// Config
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
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
