package domain

// Coffer
// Repositories
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

type RecordsRepository interface {
	WriteList(map[string][]byte)
	WriteUnsafeRecord(string, []byte)
	ReadList([]string) (map[string][]byte, error)
	DelList([]string) error

	//GetRecords([]string) ([]*Record, error) // (map[string][]byte, error)
	//SetRecords([]*Record) error             // map[string][]byte
	DelRecords([]string) error
	//SetUnsafeRecord(*Record)
	//Transaction(interface{}, map[string][]byte, *Handler) (map[string][]byte, error)
	Iterator(chan *Record)
}

type HandlersRepository interface {
	Set(string, *Handler) error
	Get(string) (*Handler, error)
}
