package domain

// Coffer
// Domain entities
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

type Handler func(interface{}, map[string][]byte) (map[string][]byte, error)

type Operation struct {
	Code byte
	Body []byte
}

type Record struct {
	Key   string
	Value []byte
}
