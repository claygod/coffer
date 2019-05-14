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
}

func New(cnf *Config) (*ResourcesControl, error) {
	m := &ResourcesControl{config: cnf}
	if m.config.DickPath != "" {
		if stat, err := os.Stat(m.config.DickPath); err != nil || !stat.IsDir() {
			return nil, fmt.Errorf("Invalid disk path: %s ", m.config.DickPath)
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
	go m.freeResourceSetter()
	return m, nil
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
	if r.config.DickPath == "" {
		return nil
	}
	us, err := disk.Usage(r.config.DickPath)
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
	if r.config.DickPath == "" {
		return true
	}
	for {
		curFree := atomic.LoadInt64(&r.freeDisk)
		if curFree-size > r.config.LimitDisk &&
			atomic.CompareAndSwapInt64(&r.freeDisk, curFree, curFree-size) {
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
		if curFree-size > r.config.LimitMemory &&
			atomic.CompareAndSwapInt64(&r.freeMemory, curFree, curFree-size) {
			return true
		} else if curFree-size <= r.config.LimitMemory {
			return false
		}
		runtime.Gosched()
	}
}

func (r *ResourcesControl) freeResourceSetter() {
	ticker := time.NewTicker(timeRefresh)
	for range ticker.C {
		r.setFreeResources()
	}
}
