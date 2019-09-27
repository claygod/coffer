package domain

// Coffer
// Repositories
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

type RecordsRepository interface {
	Reset()
	WriteList(map[string][]byte)
	WriteUnsafeRecord(string, []byte)
	ReadList([]string) (map[string][]byte, []string)
	DelListStrict([]string) []string
	DelListOptional([]string) ([]string, []string)

	//GetRecords([]string) ([]*Record, error) // (map[string][]byte, error)
	//SetRecords([]*Record) error             // map[string][]byte
	//DelRecords([]string) error
	//SetUnsafeRecord(*Record)
	//Transaction(interface{}, map[string][]byte, *Handler) (map[string][]byte, error)
	Iterator(chan *Record)
	CountRecords() int
}

type HandlersRepository interface {
	Set(string, *Handler)
	Get(string) (*Handler, error)
}
