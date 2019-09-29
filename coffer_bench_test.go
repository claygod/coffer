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

	"github.com/claygod/coffer/domain"
	"github.com/claygod/coffer/reports/codes"
	//"github.com/claygod/coffer/services/journal"
	//"github.com/claygod/coffer/services/resources"
	//"github.com/claygod/coffer/usecases"
)

var keyConcurent int64

func BenchmarkClean(b *testing.B) {
	forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferFast(b, 1000, 1000, 100, 1000) //createAndStartNewCofferLengthB(b, 10, 100)
	if err != nil {
		b.Error(err)
		return
	}
	defer forTestClearDir(dirPath)
	defer cof1.Stop()
	defer forTestClearDir(dirPath)
}

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
	//fmt.Println("111_Запущена копия бенчмарка")
	b.StopTimer()
	forTestClearDir(dirPath)
	//time.Sleep(1 * time.Second)
	//fmt.Println("====================Parallel======================")
	cof1, err := createAndStartNewCofferFast(b, 1000, 1000, 100, 1000) //createAndStartNewCofferLengthB(b, 10, 100)
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
	//fmt.Println("222_Запущена копия бенчмарка")
	b.StopTimer()
	forTestClearDir(dirPath)
	//time.Sleep(1 * time.Second)
	//fmt.Println("====================Parallel======================")
	cof1, err := createAndStartNewCofferFast(b, 1000, 1000, 100, 1000) //  createAndStartNewCofferLengthB(b, 10, 100)
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

func BenchmarkCofferTransactionSequence32HiConcurent(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
	//fmt.Println("333_Запущена копия бенчмарка")
	b.StopTimer()
	forTestClearDir(dirPath)
	cof10, err := createAndStartNewCofferFast(b, 10, 1000, 100, 1000) //  createAndStartNewCofferLengthB(b, 10, 100)
	if err != nil {
		b.Error(err)
		return
	}
	defer forTestClearDir(dirPath)
	defer cof10.Stop()
	defer forTestClearDir(dirPath)

	for x := 0; x < 500; x += 1 {
		key := strconv.Itoa(x)
		rep := cof10.Write(key, []byte(key))
		if rep.Code >= codes.Warning {
			b.Error(fmt.Sprintf("Code_: %d , err: %v", rep.Code, rep.Error))
		}
	}
	//fmt.Println("DB filled", cof10.Count())
	atomic.AddInt64(&keyConcurent, 100)
	cof10.ReadList([]string{"101", "102"})
	//fmt.Println(rep)
	//b.SetParallelism(32)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rep := cof10.Transaction("exchange", []string{"101", "102"}, nil)
		if rep.Code >= codes.Warning || rep.Error != nil {
			b.Error("EEEEEEEEERRRRRRRRRRRRRRRRR")
			//b.Error(fmt.Sprintf("Code: %d , key1: %s, key2: %s . Err %v", rep.Code, keys[0], keys[1], rep.Error))
		}
		//fmt.Println(rep)
	}
}

func BenchmarkCofferTransactionPar32NotConcurent(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
	//fmt.Println("444_Запущена копия бенчмарка")
	b.StopTimer()
	forTestClearDir(dirPath)
	cof11, err := createAndStartNewCofferFast(b, 1000, 10000, 500, 1000) //  createAndStartNewCofferLengthB(b, 10, 100)
	if err != nil {
		b.Error(err)
		return
	}
	defer forTestClearDir(dirPath)
	defer cof11.Stop()
	defer forTestClearDir(dirPath)

	for x := 0; x < 70000; x += 100 {
		list := make(map[string][]byte, 100)
		for z := x; z < x+100; z++ {
			key := strconv.Itoa(z)
			list[key] = []byte("a" + key + "b")
		}
		rep := cof11.WriteList(list)
		if rep.Code >= codes.Warning {
			b.Error(fmt.Sprintf("Code_: %d , err: %v", rep.Code, rep.Error))
		}
	}

	// for x := 0; x < 70000; x += 1 {
	// 	key := strconv.Itoa(x)
	// 	rep := cof10.Write(key, []byte(key))
	// 	if rep.Code >= codes.Warning {
	// 		b.Error(fmt.Sprintf("Code_: %d , err: %v", rep.Code, rep.Error))
	// 	}
	// }
	//fmt.Println("DB filled", cof11.Count())
	atomic.AddInt64(&keyConcurent, 100)
	cof11.ReadList([]string{"101", "102"})
	//fmt.Println(rep)
	b.SetParallelism(32)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			u1 := int64(uint16(atomic.AddInt64(&keyConcurent, 1)))
			u2 := int64(uint16(atomic.AddInt64(&keyConcurent, 1)))
			atomic.AddInt64(&keyConcurent, 100)
			//tStart := time.Now().UnixNano()

			rep := cof11.Transaction("exchange", []string{strconv.FormatInt(u1, 10), strconv.FormatInt(u2, 10)}, nil)
			if rep.Code >= codes.Warning {
				b.Error(fmt.Sprintf("Code: %d , key1: %d, key2: %d", rep.Code, u1, u2))
			}
			//fmt.Println(rep)
			//fmt.Println("Время проведения оперерации ", time.Now().UnixNano()-tStart, u1, u2)
		}
	})
}

func BenchmarkCofferTransactionPar32HalfConcurent(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
	//fmt.Println("555_Запущена копия бенчмарка")
	b.StopTimer()
	forTestClearDir(dirPath)
	cof12, err := createAndStartNewCofferFast(b, 1000, 10000, 500, 1000) //  createAndStartNewCofferLengthB(b, 10, 100)
	if err != nil {
		b.Error(err)
		return
	}
	defer forTestClearDir(dirPath)
	defer cof12.Stop()
	defer forTestClearDir(dirPath)

	for x := 0; x < 70000; x += 100 {
		list := make(map[string][]byte, 100)
		for z := x; z < x+100; z++ {
			key := strconv.Itoa(z)
			list[key] = []byte("a" + key + "b")
		}
		rep := cof12.WriteList(list)
		if rep.Code >= codes.Warning {
			b.Error(fmt.Sprintf("Code_: %d , err: %v", rep.Code, rep.Error))
		}
	}

	// for x := 0; x < 70000; x += 1 {
	// 	key := strconv.Itoa(x)
	// 	rep := cof10.Write(key, []byte(key))
	// 	if rep.Code >= codes.Warning {
	// 		b.Error(fmt.Sprintf("Code_: %d , err: %v", rep.Code, rep.Error))
	// 	}
	// }
	//fmt.Println("DB filled", cof12.Count())
	atomic.AddInt64(&keyConcurent, 100)
	cof12.ReadList([]string{"101", "102"})
	//fmt.Println(rep)
	b.SetParallelism(32)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			u1 := int64(uint16(atomic.AddInt64(&keyConcurent, 0)))
			u2 := int64(uint16(atomic.AddInt64(&keyConcurent, 1)))
			atomic.AddInt64(&keyConcurent, 100)
			//tStart := time.Now().UnixNano()

			rep := cof12.Transaction("exchange", []string{strconv.FormatInt(u1, 10), strconv.FormatInt(u2, 10)}, nil)
			if rep.Code >= codes.Warning {
				b.Error(fmt.Sprintf("Code: %d , key1: %d, key2: %d", rep.Code, u1, u2))
			}
			//fmt.Println(rep)
			//fmt.Println("Время проведения оперерации ", time.Now().UnixNano()-tStart, u1, u2)
		}
	})
}

// =======================================================================
// =========================== HELPERS ===================================
// =======================================================================

func createAndStartNewCofferFast(t *testing.B, batchSize int, limitRecordsPerLogfile int, maxKeyLength int, maxValueLength int) (*Coffer, error) {
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

func createNewCofferFast(batchSize int, limitRecordsPerLogfile int, maxKeyLength int, maxValueLength int) (*Coffer, error, error) {
	// jCnf := &journal.Config{
	// 	BatchSize:              batchSize,
	// 	LimitRecordsPerLogfile: limitRecordsPerLogfile,
	// }
	// ucCnf := &usecases.Config{
	// 	FollowPause:             100 * time.Second, //чтобы точно не включался
	// 	LogsByCheckpoint:        1000,              //чтобы точно не включался
	// 	DirPath:                 dirPath,           // "/home/ed/goPath/src/github.com/claygod/coffer/test",
	// 	AllowStartupErrLoadLogs: true,
	// 	MaxKeyLength:            maxKeyLength,
	// 	MaxValueLength:          maxValueLength,
	// 	RemoveUnlessLogs:        true, // чистим логи после использования
	// }
	// rcCnf := &resources.Config{
	// 	LimitMemory: 1000 * megabyte, // minimum available memory (bytes)
	// 	LimitDisk:   1000 * megabyte, // minimum free disk space
	// 	DirPath:     dirPath,         // "/home/ed/goPath/src/github.com/claygod/coffer/test"
	// }

	// cnf := &Config{
	// 	JournalConfig:       jCnf,
	// 	UsecasesConfig:      ucCnf,
	// 	ResourcesConfig:     rcCnf,
	// 	MaxRecsPerOperation: 1000,
	// 	//MaxKeyLength:        100,
	// 	//MaxValueLength:      10000,
	// }
	hdlExch := domain.Handler(handlerExchange)
	return Db(dirPath).BatchSize(batchSize).
		LimitRecordsPerLogfile(limitRecordsPerLogfile).
		FollowPause(100*time.Second).
		LogsByCheckpoint(1000).
		MaxKeyLength(maxKeyLength).
		MaxValueLength(maxValueLength).
		MaxRecsPerOperation(1000000).
		Handler("exchange", &hdlExch).
		Create()

	//return New(cnf, nil)
}
