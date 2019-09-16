package batcher

// Batcher
// Batcher tests
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"os"

	//"runtime/pprof"
	"testing"
	"time"
)

func TestBatcher(t *testing.T) {
	fileName := "./test.txt"
	wr := newMockWriter(fileName)
	chIn := make(chan []byte, 100)
	batchSize := 10
	btch := NewBatcher(wr, mockAlarmHandle, chIn, batchSize)
	btch.Start()
	for u := 0; u < 25; u++ {
		chIn <- []byte{97}
	}
	time.Sleep(200 * time.Millisecond)
	wr.Close()
	f, _ := os.Open(fileName)
	st, err := f.Stat()
	if err != nil {
		t.Error("Error `stat` file")
	}
	if st.Size() != 28 {
		t.Error("Want 28, have ", st.Size())
	}

	btch.Stop()
	// os.Remove(fileName)
}

// func BenchmarkClientSequence(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
// 	b.StopTimer()
// 	fmt.Println("====================Sequence======================")
// 	clt, err := Open("./tmp.txt", 2000)
// 	if err != nil {
// 		b.Error("Error `stat` file")
// 	}
// 	//defer
// 	dummy := forTestGetDummy(100)

// 	u := 0
// 	b.SetParallelism(4)
// 	f, err := os.Create("cpu.txt")
// 	if err != nil {
// 		b.Error("could not create CPU profile: ", err)
// 	}
// 	if err := pprof.StartCPUProfile(f); err != nil {
// 		b.Error("could not start CPU profile: ", err)
// 	}
// 	b.StartTimer()
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			//fmt.Println("++++++++++++++1+++++++++", u)
// 			clt.Write(dummy)
// 			//fmt.Println("++++++++++++++2+++++++++", u)
// 			u++
// 		}
// 	})
// 	pprof.StopCPUProfile()
// 	clt.Close()
// 	// os.Remove(fileName)
// }

func BenchmarkClientParallel(b *testing.B) { // go tool pprof -web ./batcher.test ./cpu.txt
	b.StopTimer()
	fmt.Println("====================Parallel======================")
	clt, err := Open("./tmp.txt", 1000)
	if err != nil {
		b.Error("Error `stat` file")
	}
	//defer
	dummy := forTestGetDummy(100)

	u := 0
	b.SetParallelism(2)
	// f, err := os.Create("cpu.txt")
	// if err != nil {
	// 	b.Error("could not create CPU profile: ", err)
	// }
	// if err := pprof.StartCPUProfile(f); err != nil {
	// 	b.Error("could not start CPU profile: ", err)
	// }
	for x := 0; x < 256; x++ {
		//go genTraffic(clt, dummy)
	}

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			//fmt.Println("++++++++++++++1+++++++++", u)
			clt.Write(dummy)
			//fmt.Println("++++++++++++++2+++++++++", u)
			u++
		}
	})
	//pprof.StopCPUProfile()
	clt.Close()
	// os.Remove(fileName)
}

// --- Helpers for tests ---

func genTraffic(clt *Client, dummy []byte) {
	for x := 0; x < 25600000; x++ {
		clt.Write(dummy)
	}
}

type mockWriter struct {
	f *os.File
}

func newMockWriter(fileName string) *mockWriter {
	f, _ := os.Create(fileName)
	return &mockWriter{
		f: f,
	}
}
func (m *mockWriter) Write(in []byte) (int, error) {
	m.f.Write(in)
	m.f.Write([]byte("\n")) // to calculate the batch
	return len(in), nil
}
func (m *mockWriter) Close() {
	m.f.Close()
}

func mockAlarmHandle(err error) {
	panic(err)
}

func forTestGetDummy(count int) []byte {
	dummy := make([]byte, count)
	for i := 0; i < count; i++ {
		dummy[i] = 105
	}
	return dummy
}
