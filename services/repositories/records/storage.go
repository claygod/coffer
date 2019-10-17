package records

// Coffer
// Records storage
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"strings"

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

func (r *storage) readList(keys []string) (map[string][]byte, []string) {
	notFound := make([]string, 0, len(keys))
	list := make(map[string][]byte)
	for _, key := range keys {
		if value, ok := r.data[key]; ok {
			list[key] = value
		} else {
			notFound = append(notFound, key)
		}
	}
	return list, notFound
}

func (r *storage) writeList(list map[string][]byte) {
	for key, value := range list {
		r.data[key] = value
	}
}

func (r *storage) writeOne(key string, value []byte) {
	r.data[key] = value
}

func (r *storage) setOne(rec *domain.Record) {
	r.data[rec.Key] = rec.Value
}

func (r *storage) removeWhatIsPossible(keys []string) ([]string, []string) {
	removedList := make([]string, 0, len(keys))
	notFound := make([]string, 0, len(keys))
	for _, key := range keys {
		if _, ok := r.data[key]; ok {
			removedList = append(removedList, key)
			delete(r.data, key)
		} else {
			notFound = append(notFound, key)
		}
	}
	return removedList, notFound
}

func (r *storage) delAllOrNothing(keys []string) []string {
	notFound := make([]string, 0, len(keys))
	for _, key := range keys { // first check if all of these keys
		if _, ok := r.data[key]; !ok {
			notFound = append(notFound, key)
		}
	}
	if len(notFound) != 0 {
		return notFound
	}
	for _, key := range keys { // now delete
		delete(r.data, key)
	}
	return notFound
}

func (r *storage) iterator(chRecord chan *domain.Record, chFinish chan struct{}) {
	for key, value := range r.data {
		chRecord <- &domain.Record{
			Key:   key,
			Value: value,
		}
	}
	close(chFinish)
}

func (r *storage) countRecords() int {
	return len(r.data)
}

func (r *storage) keysList() []string {
	list := make([]string, 0, len(r.data))
	for key := range r.data {
		list = append(list, key)
	}
	return list
}

func (r *storage) keysListWithPrefix(prefix string) []string {
	list := make([]string, 0, len(r.data))
	for key := range r.data {
		if strings.HasPrefix(key, prefix) {
			list = append(list, key)
		}
	}
	return list
}

func (r *storage) keysListWithSuffix(suffix string) []string {
	list := make([]string, 0, len(r.data))
	for key := range r.data {
		if strings.HasSuffix(key, suffix) {
			list = append(list, key)
		}
	}
	return list
}
