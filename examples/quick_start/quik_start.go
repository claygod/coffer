package main

// Coffer
// Examples: quik start
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"

	"github.com/claygod/coffer"
	"github.com/claygod/coffer/domain"
	"github.com/claygod/coffer/examples"
)

const curDir = "./"

func main() {
	examples.ClearDir(curDir)
	defer examples.ClearDir(curDir)
	fmt.Println("Quik start: BEGIN")

	// STEP init
	hdlExch := domain.Handler(HandlerExchange)
	db, err, wrn := coffer.Db(curDir).Handler("exchange", &hdlExch).Create()
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

	// STEP write
	if rep := db.Write("foo", []byte("bar")); rep.IsCodeWarning() {
		fmt.Sprintf("Write error: code `%d` msg `%s`", rep.Code, rep.Error)
		return
	}
	if rep := db.WriteList(map[string][]byte{"john": []byte("ball"), "alice": []byte("flower")}); rep.IsCodeWarning() {
		fmt.Sprintf("Write error: code `%d` msg `%s`", rep.Code, rep.Error)
		return
	}
	fmt.Println("- write: John has ball and Alice has flower")

	// STEP transaction
	if rep := db.Transaction("exchange", []string{"john", "alice"}, nil); rep.IsCodeWarning() {
		fmt.Sprintf("Transaction error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Println("- transaction: John and Alice exchanged items")

	// STEP read
	if rep := db.Read("foo"); rep.IsCodeWarning() {
		fmt.Sprintf("Read error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	rep := db.ReadList([]string{"john", "alice"})
	if rep.IsCodeWarning() {
		fmt.Sprintf("Read error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Println(fmt.Sprintf("- read: John has %s and Alice has %s", string(rep.Data["john"]), string(rep.Data["alice"])))

	fmt.Println("Quik start: END")
}
