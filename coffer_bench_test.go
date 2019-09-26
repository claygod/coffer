package coffer

// Coffer
// API benchmarks
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/claygod/coffer/reports/codes"
	"github.com/claygod/coffer/services/journal"
	"github.com/claygod/coffer/services/resources"
	"github.com/claygod/coffer/usecases"
)

var keyConcurent int64

// func BenchmarkCofferReadParallel32HiConcurent(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
// 	fmt.Println("000Запущена копия бенчмарка")
// 	b.StopTimer()
// 	//b.SetParallelism(1)
// 	forTestClearDir(dirPath)
// 	//time.Sleep(1 * time.Second)
// 	//fmt.Println("====================Parallel======================")
// 	cof1, err := createAndStartNewCofferFast(b, 500, 100002, 100, 1000) //  createAndStartNewCofferLengthB(b, 10, 100)
// 	if err != nil {
// 		b.Error(err)
// 		return
// 	}
// 	defer cof1.Stop()
// 	defer forTestClearDir(dirPath)
// 	for x := 0; x < 100000; x += 100 {
// 		list := make(map[string][]byte, 100)
// 		for z := x; z < x+100; z++ {
// 			key := strconv.Itoa(z)
// 			list[key] = []byte("a" + key + "b")
// 		}
// 		rep := cof1.WriteList(list)
// 		if rep.Code >= codes.Warning {
// 			b.Error(fmt.Sprintf("Code_: %d , err: %v", rep.Code, rep.Error))
// 		}
// 	}
// 	fmt.Println("DB filled", cof1.Count())
// 	time.Sleep(2 * time.Second)
// 	u := 0

// 	b.StartTimer()
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			y := int(uint16(u))
// 			key := strconv.Itoa(y)
// 			rep := cof1.Read(key)
// 			if rep.Code >= codes.Warning {
// 				b.Error(fmt.Sprintf("Code: %d , key: %s", rep.Code, key))
// 			}
// 			u++
// 			//fmt.Println("++++++++", u)
// 		}
// 	})
// }

func BenchmarkCofferWriteParallel32NotConcurent(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
	b.SetParallelism(1)
	fmt.Println("111Запущена копия бенчмарка")
	b.StopTimer()
	forTestClearDir(dirPath)
	//time.Sleep(1 * time.Second)
	//fmt.Println("====================Parallel======================")
	cof1, err := createAndStartNewCofferFast(b, 500, 1000, 100, 1000) //createAndStartNewCofferLengthB(b, 10, 100)
	if err != nil {
		b.Error(err)
		return
	}
	defer cof1.Stop()
	defer forTestClearDir(dirPath)
	b.SetParallelism(32)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			u := atomic.AddInt64(&keyConcurent, 1)
			key := strconv.FormatInt(u, 10)
			rep := cof1.Write(key, []byte("aaa"+key+"bbb"))
			if rep.Code >= codes.Warning {
				b.Error(fmt.Sprintf("Code: %d , key: %s", rep.Code, key))
			}
		}
	})
}

func BenchmarkCofferWriteParallel32HiConcurent(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
	fmt.Println("222Запущена копия бенчмарка")
	b.StopTimer()
	forTestClearDir(dirPath)
	//time.Sleep(1 * time.Second)
	//fmt.Println("====================Parallel======================")
	cof1, err := createAndStartNewCofferFast(b, 500, 1000, 100, 1000) //  createAndStartNewCofferLengthB(b, 10, 100)
	if err != nil {
		b.Error(err)
		return
	}
	defer cof1.Stop()
	defer forTestClearDir(dirPath)
	u := 0
	b.SetParallelism(32)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := strconv.Itoa(u)
			rep := cof1.Write(key, []byte("aaa"+key+"bbb"))
			if rep.Code >= codes.Warning {
				b.Error(fmt.Sprintf("Code: %d , key: %s", rep.Code, key))
			}
			u++
		}
	})
}

// =======================================================================
// =========================== HELPERS ===================================
// =======================================================================

func createAndStartNewCofferFast(t *testing.B, batchSize int, limitRecordsPerLogfile int64, maxKeyLength int, maxValueLength int) (*Coffer, error) {
	cof1, err, wrn := createNewCofferFast(batchSize, limitRecordsPerLogfile, maxKeyLength, maxValueLength)
	if err != nil {
		return nil, err
	} else if wrn != nil {
		t.Log(wrn)
	}
	if !cof1.Start() {
		return nil, fmt.Errorf("Failed to start (cof)")
	}
	return cof1, nil
}

func createNewCofferFast(batchSize int, limitRecordsPerLogfile int64, maxKeyLength int, maxValueLength int) (*Coffer, error, error) {
	jCnf := &journal.Config{
		BatchSize:              batchSize,
		LimitRecordsPerLogfile: limitRecordsPerLogfile,
	}
	ucCnf := &usecases.Config{
		FollowPause:             100 * time.Second, //чтобы точно не включался
		LogsByCheckpoint:        1000,              //чтобы точно не включался
		DirPath:                 dirPath,           // "/home/ed/goPath/src/github.com/claygod/coffer/test",
		AllowStartupErrLoadLogs: true,
		MaxKeyLength:            maxKeyLength,
		MaxValueLength:          maxValueLength,
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
		MaxRecsPerOperation: 1000,
		//MaxKeyLength:        100,
		//MaxValueLength:      10000,
	}
	return New(cnf)
}
