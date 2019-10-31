package main

// Coffer
// Examples: finance
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/claygod/coffer"
	"github.com/claygod/coffer/domain"
	"github.com/claygod/coffer/examples"
)

const curDir = "./"

func main() {
	examples.ClearDir(curDir)
	defer examples.ClearDir(curDir)
	fmt.Printf("Financial example:\n\n")

	// STEP init
	hdlCredit := domain.Handler(HandlerCredit)
	hdlDebit := domain.Handler(HandlerDebit)
	hdlTransfer := domain.Handler(HandlerTransfer)
	hdlMultiTransfer := domain.Handler(HandlerMultiTransfer)
	db, err, wrn := coffer.Db(curDir).Handler("credit", &hdlCredit).Handler("debit", &hdlDebit).Handler("transfer", &hdlTransfer).Handler("multi_transfer", &hdlMultiTransfer).Create()
	switch {
	case err != nil:
		fmt.Println("Error:", err)
		return
	case wrn != nil:
		fmt.Println("Warning:", err)
		return
	case !db.Start():
		fmt.Println("Error: not start")
		return
	}
	defer db.Stop()

	//STEP create accounts
	if rep := db.WriteList(map[string][]byte{"john_usd": uint64ToBytes(90), "alice_usd": uint64ToBytes(5), "john_stock": uint64ToBytes(5), "alice_stock": uint64ToBytes(25)}, true); rep.IsCodeError() {
		fmt.Printf("Write error: code `%d` msg `%s`", rep.Code, rep.Error)
		return
	}

	//STEP initial
	rep := db.ReadList([]string{"john_usd", "alice_usd", "john_stock", "alice_stock"})
	if rep.IsCodeError() {
		fmt.Sprintf("Read error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Printf(" initial: John's usd=%d stocks=%d  Alice's usd=%d stocks=%d\n",
		bytesToUint64(rep.Data["john_usd"]),
		bytesToUint64(rep.Data["john_stock"]),
		bytesToUint64(rep.Data["alice_usd"]),
		bytesToUint64(rep.Data["alice_stock"]))
	showState(db)

	//STEP credit
	if rep := db.Transaction("credit", []string{"john_usd"}, uint64ToBytes(5)); rep.IsCodeError() {
		fmt.Printf("Transaction error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Println(" credit: 5 usd were withdrawn from John’s account")
	showState(db)

	//STEP debit
	if rep := db.Transaction("debit", []string{"alice_stock"}, uint64ToBytes(2)); rep.IsCodeError() {
		fmt.Printf("Transaction error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Println(" debit: two stocks added to Alice's account")
	showState(db)

	//STEP transfer
	req1 := ReqTransfer{From: "john_usd", To: "alice_usd", Amount: 3}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(req1); err != nil {
		fmt.Println(err)
		return
	}
	if rep := db.Transaction("transfer", []string{"john_usd", "alice_usd"}, buf.Bytes()); rep.IsCodeError() {
		fmt.Printf("Transaction error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Println(" transfer: John transferred Alice $ 3")
	showState(db)

	//STEP purchase/sale
	t1 := ReqTransfer{From: "john_usd", To: "alice_usd", Amount: 50}
	t2 := ReqTransfer{From: "alice_stock", To: "john_stock", Amount: 5}
	req2 := []ReqTransfer{t1, t2}
	buf.Reset()
	enc = gob.NewEncoder(&buf)
	if err := enc.Encode(req2); err != nil {
		fmt.Println(err)
		return
	}
	if rep := db.Transaction("multi_transfer", []string{"john_usd", "alice_usd", "john_stock", "alice_stock"}, buf.Bytes()); rep.IsCodeError() {
		fmt.Printf("Transaction error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Println(" purchase/sale: John bought 5 stocks from Alice for $ 50")
	showState(db)
}

func showState(db *coffer.Coffer) {
	rep := db.ReadList([]string{"john_usd", "alice_usd", "john_stock", "alice_stock"})
	if rep.IsCodeError() {
		fmt.Sprintf("Read error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Printf("-------------------------------------------------------------->>>  John's usd:%3d stocks:%3d  Alice's usd:%3d stocks:%3d\n",
		bytesToUint64(rep.Data["john_usd"]),
		bytesToUint64(rep.Data["john_stock"]),
		bytesToUint64(rep.Data["alice_usd"]),
		bytesToUint64(rep.Data["alice_stock"]))
}
