package journal

// Coffer
// Journal (tests)
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"log"
	"os"
	"runtime/pprof"
	"testing"

	//"time"

	"github.com/claygod/tools/batcher"
)

func BenchmarkClient(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
	b.StopTimer()
	clt, err := batcher.Open("./tmp.txt", 2000)
	if err != nil {
		b.Error("Error `stat` file")
	}
	defer clt.Close()
	dummy := forTestGetDummy(100)

	u := 0
	b.SetParallelism(256)
	f, err := os.Create("cpu.txt")
	if err != nil {
		b.Error("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		b.Error("could not start CPU profile: ", err)
	}
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			clt.Write(dummy)
			u++
		}
	})
	pprof.StopCPUProfile()

	// os.Remove(fileName)
}

func BenchmarkBatcherClient(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
	b.StopTimer()
	//time.Sleep(10 * time.Millisecond)
	clt, err := batcher.Open("./tmp.txt", 2000)
	if err != nil {
		b.Error("Error `stat` file")
	}
	defer clt.Close()
	dummy := forTestGetDummy(100)

	u := 0
	b.SetParallelism(256)
	// f, err := os.Create("cpu.txt")
	// if err != nil {
	// 	b.Error("could not create CPU profile: ", err)
	// }
	// if err := pprof.StartCPUProfile(f); err != nil {
	// 	b.Error("could not start CPU profile: ", err)
	// }
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			clt.Write(dummy)
			u++
		}
	})
	//pprof.StopCPUProfile()

	// os.Remove(fileName)
}

func BenchmarkNew1(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
	b.StopTimer()
	j := New("./", forTestAlarmer, nil, 2000)
	defer j.client.Close()
	dummy := forTestGetDummy(100)
	u := 0
	b.SetParallelism(256)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			j.client.Write(dummy)
			u++
		}
	})
	pprof.StopCPUProfile()

	// os.Remove(fileName)
}

func BenchmarkNew2(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
	b.StopTimer()
	j := New("./", forTestAlarmer, nil, 2000)
	defer j.client.Close()
	dummy := forTestGetDummy(100)
	u := 0
	b.SetParallelism(256)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			j.Write(dummy)
			u++
		}
	})
	pprof.StopCPUProfile()

	// os.Remove(fileName)
}

func forTestGetDummy(count int) []byte {
	dummy := make([]byte, count)
	for i := 0; i < count; i++ {
		dummy[i] = 105
	}
	return dummy
}

func forTestAlarmer(err error) {
	log.Println(err)
}
