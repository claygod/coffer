package batcher

// Batcher
// Client
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"log"
	"os"
	"runtime"
)

/*
Client - for ease of use of the batcher in the typis case.
*/
type Client struct {
	b *batcher
	f *os.File
	//chIn chan []byte
}

/*
Open - client creation and batcher.
*/
func Open(filePath string, batchSize int, alarmFunc func(error)) (*Client, error) {
	//fmt.Println("Создание нового клиента ", filePath)
	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	//chIn := make(chan []byte, batchSize)
	nb := newBatcher(newWriter(f), alarmFunc, batchSize)
	nb.start()

	return &Client{
		b: nb,
		f: f,
		//chIn: chIn,
	}, nil
}

/*
Write - write data.
*/
func (c *Client) Write(in []byte) {

	//fmt.Println("step 1", string(in))
	c.b.chInput <- in
	//fmt.Println("step 2")

	//fmt.Println("step 3")
	ch := c.b.getChan()
	//fmt.Println("step 4")
	<-ch
	//fmt.Println("step 5")
}

/*
Close - stop and close Client.
*/
func (c *Client) Close() {
	if c.b == nil {
		return
	}
	c.b.stop()
	for {
		if len(c.b.chInput) == 0 {
			return
		}
		runtime.Gosched()
	}
}

/*
alarm - errors log.
*/
func alarm(err error) {
	log.Print(err)
}

/*
writer - intermediate structure for writing to file.
*/
type writer struct {
	f *os.File
}

/*
newWriter - create new filewriter.
*/
func newWriter(f *os.File) *writer {
	return &writer{
		f: f,
	}
}

/*
Write - write data to a file with synchronization
*/
func (w *writer) Write(in []byte) (int, error) {
	i, err := w.f.Write(in)
	//i, err = w.f.Write([]byte{99}) //TODO: зачем тут ещё один символ??
	if err == nil {
		err = w.f.Sync()
	}
	return i, err
}
