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
Хранилище store вынесено в отдельную сущность для того, чтобы при необходимости
можно было бы работать не с одним хранилищем, а с их массивом.
*/
type Records struct {
	mtx   sync.RWMutex
	store *storage
}

func New() *Records {
	return &Records{
		store: newStorage(),
	}
}

func (s *Records) GetRecords(keys []string) ([]*domain.Record, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.store.get(keys)
}

func (s *Records) SetRecords(in []*domain.Record) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.store.set(in)
}

func (s *Records) DelRecords(keys []string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.store.del(keys)
}

func (s *Records) SetUnsafeRecord(rec *domain.Record) {
	s.store.setOne(rec)
}

// func (s *Records) transaction(v interface{}, curValues map[string][]byte, f *domain.Handler) (map[string][]byte, error) {
// 	s.mtx.RLock()
// 	defer s.mtx.RUnlock()
// 	return f(v, curValues)
// }

func (s *Records) Iterator(chRecord chan *domain.Record) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	chFinish := make(chan struct{})
	s.store.iterator(chRecord, chFinish)
	<-chFinish
	close(chRecord)
}
