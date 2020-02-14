package records

// Coffer
// In-Memory records repository
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"sync"

	"github.com/claygod/coffer/domain"
)

/*
Records - in-memory data repository.
The store is moved to a separate entity so that, if necessary
It would be possible to work not with one repository, but with their array.
*/
type Records struct {
	mtx   sync.RWMutex
	store *storage
}

/*
New - create new Records.
*/
func New() *Records {
	return &Records{
		store: newStorage(),
	}
}

/*
Reset - create new storage.
*/
func (s *Records) Reset() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.store = newStorage()
}

/*
WriteList - write records list.
*/
func (s *Records) WriteList(list map[string][]byte) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.store.writeList(list)
}

/*
WriteListStrict - write records list (strict mode).
*/
func (s *Records) WriteListStrict(list map[string][]byte) []string {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.store.writeListIfNot(list)
}

/*
WriteListOptional - write records list (strict mode).
*/
func (s *Records) WriteListOptional(list map[string][]byte) []string {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.store.writeListOptional(list)
}

/*
WriteListUnsafe - unsafe write records list.
*/
func (s *Records) WriteListUnsafe(list map[string][]byte) {
	s.store.writeList(list)
}

/*
ReadList - read records list.
*/
func (s *Records) ReadList(list []string) (map[string][]byte, []string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.store.readList(list)
}

/*
ReadListUnsafe - unsafe read records list.
*/
func (s *Records) ReadListUnsafe(list []string) (map[string][]byte, []string) {
	return s.store.readList(list)
}

/*
DelListStrict - remove records list (strict mode).
*/
func (s *Records) DelListStrict(keys []string) []string {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.store.delAllOrNothing(keys)
}

/*
DelListOptional - remove records list (optional mode).
*/
func (s *Records) DelListOptional(keys []string) ([]string, []string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.store.removeWhatIsPossible(keys)
}

/*
Iterator - required when saving to file.
*/
func (s *Records) Iterator(chRecord chan *domain.Record) { // required when saving to file - требуется при сохранении в файл
	s.mtx.Lock()
	defer s.mtx.Unlock()
	chFinish := make(chan struct{})
	s.store.iterator(chRecord, chFinish)
	<-chFinish
	close(chRecord)
}

/*
CountRecords - total count records.
*/
func (s *Records) CountRecords() int {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.store.countRecords()
}

/*
RecordsList - get total keys list.
*/
func (s *Records) RecordsList() []string {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.store.keysList()
}

/*
RecordsListWithPrefix - get total keys list with prefix.
*/
func (s *Records) RecordsListWithPrefix(prefix string) []string {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.store.keysListWithPrefix(prefix)
}
