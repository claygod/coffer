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
	fmt.Println("Financial example: BEGIN")

	// STEP init
	//hdlNewAccount := domain.Handler(HandlerNewAccount)
	hdlCredit := domain.Handler(HandlerCredit)
	hdlDebit := domain.Handler(HandlerDebit)
	hdlTransfer := domain.Handler(HandlerTransfer)
	//hdlMultiTransfer := domain.Handler(HandlerMultiTransfer)
	db, err, wrn := coffer.Db(curDir).
		//Handler("new_account", &hdlNewAccount).
		Handler("credit", &hdlCredit).
		Handler("debit", &hdlDebit).
		Handler("transfer", &hdlTransfer).
		//Handler("multi_transfer", &hdlMultiTransfer).
		Create()
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
	fmt.Println("- init: db created and started")

	//STEP create accounts
	if rep := db.WriteList(map[string][]byte{"john_usd": uint64ToBytes(100), "alice_usd": uint64ToBytes(5), "john_stock": uint64ToBytes(5), "alice_stock": uint64ToBytes(25)}); rep.IsCodeWarning() {
		fmt.Printf("Write error: code `%d` msg `%s`", rep.Code, rep.Error)
		return
	}
	fmt.Println("- accounts: created")

	//STEP credit
	if rep := db.Transaction("credit", []string{"john_usd"}, uint64(5)); rep.IsCodeWarning() {
		fmt.Printf("Transaction error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Println("- credit: 5 usd were withdrawn from John’s account")

	//STEP debit
	if rep := db.Transaction("debit", []string{"alice_stock"}, uint64(2)); rep.IsCodeWarning() {
		fmt.Printf("Transaction error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Println("- debit: two stocks added to Alice's account")

	//STEP transfer
	req1 := ReqTransfer{From: "john_usd", To: "alice_usd", Amount: 3}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(req1); err != nil {
		fmt.Println(err)
		return
	}
	if rep := db.Transaction("transfer", []string{"john_usd", "alice_usd"}, buf.Bytes()); rep.IsCodeWarning() {
		fmt.Printf("Transaction error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Println("- transfer: John transferred Alice $ 3")

	//STEP purchase/sale
	// t1 := &ReqTransfer{From: "john_usd", To: "alice_usd", Amount: 50}
	// t2 := &ReqTransfer{From: "alice_stock", To: "john_stock", Amount: 5}
	// req2 := []*ReqTransfer{t1, t2}
	// if rep := db.Transaction("multi_transfer", []string{"john_usd", "alice_usd", "john_stock", "alice_stock"}, req2); rep.IsCodeWarning() {
	// 	fmt.Printf("Transaction error: code `%v` msg `%v`", rep.Code, rep.Error)
	// 	return
	// }
	// fmt.Println("- purchase/sale: John bought 5 stocks from Alice for $ 50")

	fmt.Println("Financial example: END")
}
