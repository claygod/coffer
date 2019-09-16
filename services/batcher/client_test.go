package batcher

// Batcher
// Client tests
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"os"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	fileName := "./test2.txt"
	bc, err := Open(fileName, 5)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 7; i++ {
		go bc.Write([]byte{97})
	}
	time.Sleep(1000 * time.Millisecond)
	//bc.Close()
	//os.Remove(fileName)
}

// --- Helpers for tests ---
