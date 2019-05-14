package coffer

// Coffer
// Config
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

const (
	stateStopped int64 = iota
	stateStarted
	statePanic
)

const (
	maxKeyLength   int   = int(uint64(1)<<16) - 1
	maxValueLength int   = int(uint64(1)<<48) - 1
	megabyte       int64 = 1024 * 1024
)

const (
	codeWriteList byte = iota //codeWrite
	codeTransaction
	codeDeleteList
)
