package usecases

// Coffer
// Records interactor
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"github.com/claygod/coffer/domain"
)

type RecordsInteractor struct {
}

func (r *RecordsInteractor) GetRecords([]string) ([]*domain.Record, error) { // (map[string][]byte, error)
	return nil, nil
}

func (r *RecordsInteractor) SetRecords([]*domain.Record) error { // map[string][]byte
	return nil
}
func (r *RecordsInteractor) DelRecords([]string) error {
	return nil
}
func (r *RecordsInteractor) SetUnsafeRecord(*domain.Record) error {
	return nil
}
func (r *RecordsInteractor) Transaction(interface{}, map[string][]byte, *domain.Handler) (map[string][]byte, error) {
	return nil, nil
}
