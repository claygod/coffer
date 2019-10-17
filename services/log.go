package services

// Coffer
// Log to out
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"log"
)

/*
Log -logger for joutnal.
*/
type Log struct {
	prefix string
}

/*
NewLog - create new Log
*/
func NewLog(prefix string) *Log {
	return &Log{
		prefix: prefix,
	}
}

/*
Write - write to log.
*/
func (l *Log) Write(in []byte) (int, error) {
	go log.Print(string(in))
	return len(in), nil
}
