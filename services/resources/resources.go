package resources

// Coffer
// Resources API
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/claygod/coffer/services/startstop"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

/*
ResourcesControl - indicator of the status of the physical memory (and disk) of the device.
if DiskPath == "" in config, then free disk space we do not control.
*/
type ResourcesControl struct {
	config     *Config
	freeMemory int64
	freeDisk   int64
	counter    int64
	starter    *startstop.StartStop
	//hasp       int64
}

/*
New - create ResourcesControl
*/
func New(cnf *Config) (*ResourcesControl, error) {
	m := &ResourcesControl{
		config:  cnf,
		starter: startstop.New(),
	}
	if m.config.DirPath != "" {
		if stat, err := os.Stat(m.config.DirPath); err != nil || !stat.IsDir() {
			return nil, fmt.Errorf("Invalid disk path: %s ", m.config.DirPath)
		}
	}
	if err := m.setFreeResources(); err != nil {
		return nil, err
	}
	if m.freeDisk < m.config.LimitDisk*2 {
		return nil, fmt.Errorf("Low available disk: %d bytes", m.freeDisk)
	}
	if m.freeMemory < m.config.LimitMemory*2 {
		return nil, fmt.Errorf("Low available memory: %d bytes", m.freeMemory)
	}
	//m.starter.Start()
	//atomic.StoreInt64(&m.hasp, 1)
	//m.setFreeResources()
	//go m.freeResourceSetter()
	return m, nil
}

/*
Start - launch ResourcesControl
*/
func (r *ResourcesControl) Start() bool {
	res := r.starter.Start()
	if res {
		r.setFreeResources()
		go r.freeResourceSetter()
	}
	return res
	//atomic.StoreInt64(&r.hasp, 0)
}

func (r *ResourcesControl) Stop() bool {
	return r.starter.Stop()
	//atomic.StoreInt64(&r.hasp, 0)
}

/*
GetPermission - get permission to use memory (and disk).
*/
func (r *ResourcesControl) GetPermission(size int64) bool {
	counterNew := atomic.AddInt64(&r.counter, 1)
	if int8(counterNew) == 0 {
		r.setFreeResources()
	}
	if r.getPermissionMemory(size) && r.getPermissionDisk(size) {
		return true
	}
	return false
}

func (r *ResourcesControl) setFreeResources() error {
	if err := r.setFreeDisk(); err != nil {
		return err
	}
	if err := r.setFreeMemory(); err != nil {
		return err
	}
	return nil
}

func (r *ResourcesControl) setFreeDisk() error {
	if r.config.DirPath == "" {
		return nil
	}
	us, err := disk.Usage(r.config.DirPath)
	if err != nil {
		atomic.StoreInt64(&r.freeDisk, 0)
		return err
	} else {
		atomic.StoreInt64(&r.freeDisk, int64(us.Free))
		return nil
	}
}

func (r *ResourcesControl) setFreeMemory() error {
	vms, err := mem.VirtualMemory()
	if err != nil {
		atomic.StoreInt64(&r.freeMemory, 0)
		return err
	} else {
		atomic.StoreInt64(&r.freeMemory, int64(vms.Available))
		return nil
	}
}

func (r *ResourcesControl) getPermissionDisk(size int64) bool {
	if r.config.DirPath == "" {
		return true
	}
	for {
		curFree := atomic.LoadInt64(&r.freeDisk)
		//fmt.Println("R:R:curFree: ", curFree, size, r.config.LimitDisk)
		if curFree-size > r.config.LimitDisk &&
			atomic.CompareAndSwapInt64(&r.freeDisk, curFree, curFree-size) {
			//fmt.Println("R:R:curFree: ", true)
			return true
		} else if curFree-size <= r.config.LimitDisk {
			return false
		}
		runtime.Gosched()
	}
}

func (r *ResourcesControl) getPermissionMemory(size int64) bool {
	for {
		curFree := atomic.LoadInt64(&r.freeMemory)
		//fmt.Println("R:M:curFree: ", curFree, size, r.config.LimitMemory)
		if curFree-size > r.config.LimitMemory &&
			atomic.CompareAndSwapInt64(&r.freeMemory, curFree, curFree-size) {
			//fmt.Println("R:M:curFree: ", true)
			return true
		} else if curFree-size <= r.config.LimitMemory {
			return false
		}
		runtime.Gosched()
	}
}

func (r *ResourcesControl) freeResourceSetter() {
	var counter int64
	ticker := time.NewTicker(timeRefresh)
	for range ticker.C {
		if r.starter.IsReady() {
			return
		}
		// if atomic.LoadInt64(&m.hasp) == 0 {
		// 	return
		// }
		if !r.starter.Add() {
			return
		}
		if byte(counter) == 0 {
			r.setFreeResources()
		}
		r.starter.Done()
	}
}
