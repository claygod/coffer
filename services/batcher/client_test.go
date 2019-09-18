package batcher

// Batcher
// Client tests
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"os"
	//"fmt"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	fileName := "./test2.txt"
	bc, err := Open(fileName, 5, alarm)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 7; i++ {
		go bc.Write([]byte{97})
		//fmt.Println(i)
	}
	time.Sleep(1000 * time.Millisecond)
	bc.Close()

	f, _ := os.Open(fileName)
	st, err := f.Stat()
	if err != nil {
		t.Error("Error `stat` file")
	}
	if st.Size() != 7 {
		t.Error("Want 7, have ", st.Size())
	}
	//os.Remove(fileName)
}

// --- Helpers for tests ---
