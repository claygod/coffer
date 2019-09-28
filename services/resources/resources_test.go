package resources

// Coffer
// Resources Tests
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"os"
	"runtime"
	"strconv"
	"testing"
)

const overReq int64 = 1000000000000000000

var badPathWin string = "c:\\qwertyzzzzzzzzzz"
var badPathNix string = "/qwertyzzzzzzzzzzzzz"

func TestGenBadPath(t *testing.T) {
	for i := 0; i < 100000000000; i++ {
		path := ""
		if runtime.GOOS == "windows" {
			path = "c:\\" + strconv.Itoa(i)
		} else {
			path = "/" + strconv.Itoa(i)
		}
		if stat, err := os.Stat(path); err != nil || !stat.IsDir() {
			if runtime.GOOS == "windows" {
				badPathWin = path
			} else {
				badPathNix = path
			}
			break
		}
	}
}

func TestGetPermissionWithoutDiskLimit100(t *testing.T) {
	cnf := &Config{
		LimitMemory: 100,
		LimitDisk:   100,
		DirPath:     "",
	}
	if runtime.GOOS == "windows" {
		cnf.DirPath = "c:\\"
	} else {
		cnf.DirPath = "/"
	}

	r, err := New(cnf)
	if err != nil {
		t.Error(err)
	}
	if !r.GetPermission(1) {
		t.Error("Could not get permission with minimum requirements")
	}
	if r.GetPermission(overReq) {
		t.Error("Permission received for too large requirements")
	}
}

func TestGetPermissionWithoutDiskLimit10000000000(t *testing.T) {
	cnf := &Config{
		LimitMemory: 100,
		LimitDisk:   1000000000000,
		DirPath:     "",
	}
	_, err := New(cnf)
	if err == nil {
		t.Error("Permission received for too large limit")
	}
}

func TestGetPermissionWithoutMemoryLimit10000000000(t *testing.T) {
	cnf := &Config{
		LimitMemory: 1000000000000,
		LimitDisk:   100,
		DirPath:     "",
	}
	_, err := New(cnf)
	if err == nil {
		t.Error("Permission received for too large limit")
	}
}

func TestGetPermissionWithDisk(t *testing.T) {
	cnf := &Config{
		LimitMemory: 100,
		//AddRatioMemory: 5,
		LimitDisk: 100,
		//AddRatioDisk:   5,
	}
	if runtime.GOOS == "windows" {
		cnf.DirPath = "c:\\"
	} else {
		cnf.DirPath = "/"
	}
	r, err := New(cnf)
	if err != nil {
		t.Error(err)
	}
	if !r.GetPermission(1) {
		t.Error("Could not get permission with minimum requirements")
	}
	if r.GetPermission(overReq) {
		t.Error("Permission received for too large requirements")
	}
}

func TestGetPermissionWithDiskBadPath(t *testing.T) {
	cnf := &Config{
		LimitMemory: 100,
		//AddRatioMemory: 5,
		LimitDisk: 100,
		//AddRatioDisk:   5,
	}
	if runtime.GOOS == "windows" {
		cnf.DirPath = badPathWin
	} else {
		cnf.DirPath = badPathNix
	}
	_, err := New(cnf)
	if err == nil {
		t.Errorf("Wrong path %s should have caused an error", cnf.DirPath)
	}
}

func BenchmarkSetFreeMemory(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
	b.StopTimer()
	cnf := &Config{
		LimitMemory: 100,
		//AddRatioMemory: 5,
		LimitDisk: 100,
		//AddRatioDisk:   5,
	}
	if runtime.GOOS == "windows" {
		cnf.DirPath = "c:\\"
	} else {
		cnf.DirPath = "/"
	}
	r, err := New(cnf)
	if err != nil {
		b.Error(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		r.setFreeMemory()
	}
}

func BenchmarkSetFreeDisk(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
	b.StopTimer()
	cnf := &Config{
		LimitMemory: 100,
		//AddRatioMemory: 5,
		LimitDisk: 100,
		//AddRatioDisk:   5,
	}
	if runtime.GOOS == "windows" {
		cnf.DirPath = "c:\\"
	} else {
		cnf.DirPath = "/"
	}
	r, err := New(cnf)
	if err != nil {
		b.Error(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		r.setFreeDisk()
	}
}
