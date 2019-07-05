package services

// Coffer
// Log to out
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"log"
)

type Log struct {
	prefix string
}

func NewLog(prefix string) *Log {
	return &Log{
		prefix: prefix,
	}
}

func (l *Log) Write(in []byte) (int, error) {
	go log.Print(string(in))
	return len(in), nil
}
