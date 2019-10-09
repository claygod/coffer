package main

// Coffer
// Examples: finance
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"

	// "github.com/claygod/coffer"
	// "github.com/claygod/coffer/domain"
	"github.com/claygod/coffer/examples"
)

const curDir = "./"

func main() {
	examples.ClearDir(curDir)
	defer examples.ClearDir(curDir)
	fmt.Println("Financial example: BEGIN")

	//TODO: add example
	// Credit/debit
	// Transfer
	// Purchase/Sale
	// Exchange

	fmt.Println("Financial example: END")
}
