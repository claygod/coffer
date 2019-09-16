package batcher

// Batcher
// Worker
// Copyright © 2018 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	//"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

/*
worker - basic cycle.

	- creates a batch
	- passes the batch to the vryter
	- check if you need to stop
	- switches the channel
	- zeroes the buffer under the new batch
*/
func (b *Batcher) worker() {
	//fmt.Println("----------22--- ")
	var buf bytes.Buffer
	for {
		//fmt.Println("-1 собрались получить с входного канала")
		var u int
		// begin
		select {
		//TODO: наполнение должно идти и в момент отправки!!
		//		case inData := <-b.chInput:
		//			if _, err := buf.Write(inData); err != nil {
		//				b.alarm(err)
		//			}
		case <-b.chStop:
			atomic.StoreInt64(&b.stopFlag, stateStop)
			return
		case inData := <-b.chInput:
			if _, err := buf.Write(inData); err != nil {
				b.alarm(err)

			} else {
				u++
			}
			//default:
			//break
		}
		// batch fill
		//fmt.Println("-2 получили, может быть ещё что-то получим")
		for i := 0; i < b.batchSize; i++ { // -1
			select {
			case inData := <-b.chInput:
				if _, err := buf.Write(inData); err != nil {
					b.alarm(err)
				} else {
					u++
				}
			default:
				break
			}
		}
		// batch to out
		bOut := buf.Bytes()
		if len(bOut) > 0 {
			//fmt.Println("************* Текущий батч - ", u)
			if _, err := b.work.Write(buf.Bytes()); err != nil {
				atomic.StoreInt64(&b.stopFlag, stateStop)
				b.alarm(err)
				return
			}
			//b.indicator.switchChan()
			//buf.Reset()
		} else {
			//fmt.Println("************** Почему-то  len(bOut) == 0 ")
			time.Sleep(100000 * time.Microsecond)
			runtime.Gosched()
		}
		// exit-check
		select {
		case <-b.chStop:
			atomic.StoreInt64(&b.stopFlag, stateStop)
			return
		default:
		}
		// cursor (indicator)  switch
		b.indicator.switchChan()
		buf.Reset()

		//time.Sleep(100 * time.Millisecond)
	}
}

func (b *Batcher) fillBuf(buf bytes.Buffer, counter *int64) error {
	// begin
	select {
	// case <-b.chStop:
	// 	atomic.StoreInt64(&b.stopFlag, stateStop)
	// 	return nil
	case inData := <-b.chInput:
		if _, err := buf.Write(inData); err != nil {
			return err

		}
		//default:
		//break
	}
	// batch fill
	//fmt.Println("-2 получили, может быть ещё что-то получим")
	for i := 0; i < b.batchSize; i++ { // -1
		select {
		case inData := <-b.chInput:
			if _, err := buf.Write(inData); err != nil {
				return err
			}
			if atomic.LoadInt64(counter) == 0 {
				break
			}
		default:
			if atomic.LoadInt64(counter) == 0 {
				break
			}
			runtime.Gosched()
		}
	}
	return nil
}

func (b *Batcher) writeBuf(buf bytes.Buffer, counter *int64) error {
	bOut := buf.Bytes()
	if len(bOut) > 0 {
		//fmt.Println("************* Текущий батч - ", u)
		if _, err := b.work.Write(buf.Bytes()); err != nil {
			// atomic.StoreInt64(&b.stopFlag, stateStop)
			// b.alarm(err)
			return err
		}
		//b.indicator.switchChan()
		//buf.Reset()
	} else {
		//fmt.Println("************** Почему-то  len(bOut) == 0 ")
		time.Sleep(100000 * time.Microsecond)
		runtime.Gosched()
	}
	atomic.AddInt64(counter, -1)
	return nil
}
