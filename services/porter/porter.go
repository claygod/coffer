package porter

// Porter
// API
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"runtime"
	"sort"
	"sync/atomic"
	"time"
)

const timePause time.Duration = 100 * time.Microsecond

const (
	stateUnlocked int32 = iota
	stateLocked
)

/*
Porter - regulates access to resources by keys.
*/
type Porter struct {
	data [(1 << 24) - 1]int32
}

func New() *Porter {
	return &Porter{}
}

/*
Catch - block certain resources. This function will infinitely try to block the necessary resources,
so if the logic of the application using this library contains errors, deadlocks, etc., this can lead to FATAL errors.
*/
func (p *Porter) Catch(keys []string) {
	hashes := p.stringsToHashes(keys)
	for i, hash := range hashes {
		if !atomic.CompareAndSwapInt32(&p.data[hash], stateUnlocked, stateUnlocked) {
			p.throw(hashes[0:i])
			runtime.Gosched()
			time.Sleep(timePause)
		}
	}
}

/*
Throw - frees access to resources. Resources MUST be blocked before this, otherwise using this library will lead to errors.
*/
func (p *Porter) Throw(keys []string) {
	p.throw(p.stringsToHashes(keys))
}

func (p *Porter) throw(hashes []int) {
	for _, hash := range hashes {
		atomic.StoreInt32(&p.data[hash], stateUnlocked)
	}
}

func (p *Porter) stringsToHashes(keys []string) []int {
	out := make([]int, 0, len(keys))
	tempArr := make(map[int]bool)
	for _, key := range keys {
		tempArr[p.stringToHashe(key)] = true
	}
	for key, _ := range tempArr {
		out = append(out, key)
	}
	sort.Ints(out)
	return out
}

func (p *Porter) stringToHashe(key string) int {
	switch len(key) {
	case 0:
		return 0
	case 1:
		return int(uint(key[0]))
	case 2:
		return int(uint(key[1])<<4) + int(uint(key[0]))
	case 3:
		return int(uint(key[2])<<8) + int(uint(key[1])<<4) + int(uint(key[0]))
	default:
		return int(uint(key[3])<<12) + int(uint(key[2])<<8) + int(uint(key[1])<<4) + int(uint(key[0]))
	}
}
