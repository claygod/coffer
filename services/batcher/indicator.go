package batcher

// Batcher
// Indicator
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const cleanerShift uint8 = 5 //128 // shift to exclude race condition

/*
indicator - the closing of the channels signals the completed task.
*/
type indicator struct {
	m      sync.Mutex
	chDone [256]chan struct{}
	cursor uint32
}

/*
newIndicator - create new Indicator.
*/
func newIndicator() *indicator {
	i := &indicator{}
	i.chDone[0] = make(chan struct{})
	for u := 0; u < 256; u++ {
		i.chDone[u] = make(chan struct{})
	}
	//go i.autoSwitcher()
	return i
}

/*
SwitchChan - switch channels:
	- a new channel is created
	- the pointer switches to a new channel
	- the old channel (with a shift) is closed
*/
func (i *indicator) switchChan() {
	i.m.Lock()
	defer i.m.Unlock()
	//fmt.Println("indicator switch ", uint8(atomic.LoadUint32(&i.cursor)))
	cursor := uint8(atomic.LoadUint32(&i.cursor))
	i.chDone[cursor+1] = make(chan struct{})
	atomic.StoreUint32(&i.cursor, uint32(cursor+1))
	//if _, ok := i.chDone[cursor-cleanerShift]; ok {
	close(i.chDone[cursor])
	//}
}

/*
getChan - get current channel.
*/
func (i *indicator) getChan() chan struct{} {
	i.m.Lock()
	defer i.m.Unlock()
	cursor := uint8(atomic.LoadUint32(&i.cursor))
	return i.chDone[cursor]
}

func (i *indicator) autoSwitcher() {
	for {
		i.switchChan()
		time.Sleep(200 * time.Microsecond)
		//fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~Автосвич!! ")
	}
}
