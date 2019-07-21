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

func (r *storage) readList(keys []string) (map[string][]byte, error) {
	var errOut error
	list := make(map[string][]byte)
	for _, key := range keys {
		if value, ok := r.data[key]; ok {
			list[key] = value
			//out = append(out, &domain.Record{Key: key, Value: value})
		} else {
			errOut = fmt.Errorf("%v %v", errOut, fmt.Errorf("Key `%s` not found", key))
		}
	}
	return list, nil
}

// func (r *storage) get(keys []string) ([]*domain.Record, error) {
// 	var errOut error
// 	out := make([]*domain.Record, 0, len(keys))
// 	for _, key := range keys {
// 		if value, ok := r.data[key]; ok {
// 			out = append(out, &domain.Record{Key: key, Value: value})
// 		} else {
// 			errOut = fmt.Errorf("%v %v", errOut, fmt.Errorf("Key `%s` not found", key))
// 		}
// 	}
// 	return out, nil
// }

func (r *storage) writeList(list map[string][]byte) {
	for key, value := range list {
		r.data[key] = value
	}
}

func (r *storage) writeOne(key string, value []byte) {
	r.data[key] = value
}

// func (r *storage) set(in []*domain.Record) {
// 	for _, rec := range in {
// 		r.data[rec.Key] = rec.Value
// 	}
// }

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
	//fmt.Println(r.data)
	for key, value := range r.data {
		//fmt.Println("++++++ ", key, value)
		chRecord <- &domain.Record{
			Key:   key,
			Value: value,
		}
	}
	close(chFinish)
}
