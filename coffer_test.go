package coffer

// Coffer
// API tests
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	// "github.com/claygod/coffer/domain"
	// "github.com/claygod/coffer/services"
	// "github.com/claygod/coffer/services/filenamer"
	// "github.com/claygod/coffer/services/journal"
	// "github.com/claygod/coffer/services/repositories/handlers"
	// "github.com/claygod/coffer/services/repositories/records"
	// "github.com/claygod/coffer/reports"
	"github.com/claygod/coffer/reports/codes"
	"github.com/claygod/coffer/services/journal"
	"github.com/claygod/coffer/services/resources"

	// "github.com/claygod/coffer/services/startstop"
	"github.com/claygod/coffer/usecases"
	// "github.com/claygod/tools/logger"
	// "github.com/claygod/tools/porter"
)

const dirPath string = "./test/"

func TestNewCoffer(t *testing.T) {
	defer forTestClearDir(dirPath)
	jCnf := &journal.Config{
		BatchSize:              2000,
		LimitRecordsPerLogfile: 5,
	}
	ucCnf := &usecases.Config{
		FollowPause:             400 * time.Millisecond,
		LogsByCheckpoint:        2,
		DirPath:                 dirPath, // "/home/ed/goPath/src/github.com/claygod/coffer/test",
		AllowStartupErrLoadLogs: true,
		MaxKeyLength:            100,
		MaxValueLength:          10000,
		RemoveUnlessLogs:        true, // чистим логи после использования
	}
	rcCnf := &resources.Config{
		LimitMemory: 1000 * megabyte, // minimum available memory (bytes)
		LimitDisk:   1000 * megabyte, // minimum free disk space
		DirPath:     dirPath,         // "/home/ed/goPath/src/github.com/claygod/coffer/test"
	}

	cnf := &Config{
		JournalConfig:       jCnf,
		UsecasesConfig:      ucCnf,
		ResourcesConfig:     rcCnf,
		MaxRecsPerOperation: 100,
		//MaxKeyLength:        100,
		//MaxValueLength:      10000,
	}
	cof, err := New(cnf)
	if err != nil {
		t.Error(err)
		return
	}
	//time.Sleep(3 * time.Second) //TODO: del
	if cof.Start() {
		defer cof.Stop()
	} else {
		t.Errorf("Failed to start (cof)")
		return
	}
	for i := 10; i < 19; i++ {
		if rep := cof.Write("aasa"+strconv.Itoa(i), []byte("bbsb")); rep.Code > codes.Warning || rep.Error != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if rep := cof.Count(); rep.Count != 9 {
		t.Errorf("Records count, have %d, wand 9.", rep.Count)
		return
	}

	b1, err := ioutil.ReadFile(dirPath + "4.log")
	if err != nil {
		t.Error(err)
		return
	}
	// специально портим один файл, и одна запись в нём при скачке должна быть потеряна
	if err := ioutil.WriteFile(dirPath+"4.log", b1[:len(b1)-2], os.ModePerm); err != nil {
		t.Error(err)
		return
	}

	cof2, err := New(cnf)
	if err != nil {
		t.Log(err)
		return
	}
	if cof2.Start() {
		defer cof2.Stop()
	} else {
		t.Errorf("Failed to start (cof2)")
		return
	}

	if rep := cof2.Count(); rep.Count != 8 { // одна запись поломана и её нет, а почему-то скачена
		t.Errorf("Records count, have %d, wand 8.", rep.Count)
		return
	}
}

func forTestClearDir(dir string) error {
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		//fmt.Println(name)
		if strings.HasSuffix(name, ".log") || strings.HasSuffix(name, ".check") || strings.HasSuffix(name, ".checkpoint") {
			os.Remove(dir + name)
		}
		//		err = os.RemoveAll(filepath.Join(dir, name))
		//		if err != nil {
		//			return err
		//		}
	}
	return nil
}
