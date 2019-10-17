package domain

// Coffer
// Repositories
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

/*
RecordsRepository - records store interface.
*/
type RecordsRepository interface {
	Reset()
	WriteList(map[string][]byte)
	WriteListUnsafe(map[string][]byte)
	//WriteUnsafeRecord(string, []byte)
	ReadList([]string) (map[string][]byte, []string)
	ReadListUnsafe([]string) (map[string][]byte, []string)
	DelListStrict([]string) []string
	DelListOptional([]string) ([]string, []string)

	Iterator(chan *Record) // требуется при сохранении в файл
	CountRecords() int
	RecordsList() []string
	RecordsListWithPrefix(string) []string
	RecordsListWithSuffix(string) []string
}

/*
HandlersRepository - handlers store interface.
*/
type HandlersRepository interface {
	Set(string, *Handler)
	Get(string) (*Handler, error)
}
