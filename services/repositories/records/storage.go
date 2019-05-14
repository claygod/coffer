package records

// Coffer
// Records storage
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"

	"github.com/claygod/coffer/domain"
)

/*
storage - easy data storage (not parallel mode).
*/
type storage struct {
	data map[string][]byte
}

func newStorage() *storage {
	return &storage{
		data: make(map[string][]byte),
	}
}

func (r *storage) get(keys []string) ([]*domain.Record, error) {
	var errOut error
	out := make([]*domain.Record, 0, len(keys))
	for _, key := range keys {
		if value, ok := r.data[key]; ok {
			out = append(out, &domain.Record{Key: key, Value: value})
		} else {
			errOut = fmt.Errorf("%v %v", errOut, fmt.Errorf("Key `%s` not found", key))
		}
	}
	return out, nil
}

func (r *storage) set(in []*domain.Record) {
	for _, rec := range in {
		r.data[rec.Key] = rec.Value
	}
}

func (r *storage) setOne(rec *domain.Record) {
	r.data[rec.Key] = rec.Value
}

func (r *storage) del(keys []string) error {
	var errOut error
	for _, key := range keys { // сначала проверяем, есть ли все эти ключи
		if _, ok := r.data[key]; !ok {
			errOut = fmt.Errorf("%v %v", errOut, fmt.Errorf("Key `%s` not found", key))
		}
	}
	if errOut != nil {
		return errOut
	}
	for _, key := range keys { // теперь удаляем
		delete(r.data, key)
	}
	return errOut
}

// func (r *storage) keys() []string { // Resource-intensive method
// 	out := make([]string, 0, len(r.data))
// 	for key, _ := range r.data {
// 		out = append(out, key)
// 	}
// 	return out
// }

// func (r *storage) len() int {
// 	return len(r.data)
// }

func (r *storage) iterator(chRecord chan *domain.Record, chFinish chan struct{}) {
	for key, value := range r.data {
		chRecord <- &domain.Record{
			Key:   key,
			Value: value,
		}
	}
	close(chFinish)
}
