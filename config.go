package coffer

// Coffer
// Config
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

type Config struct {
	//DataPath            string
	MaxRecsPerOperation int
	MaxKeyLength        int
	MaxValueLength      int
}

const (
	stateStopped int64 = iota
	stateStarted
	statePanic
)

// const (
// 	megabyte int64 = 1024 * 1024
// )

const (
	codeWriteList byte = iota //codeWrite
	codeTransaction
	codeDeleteList
)
