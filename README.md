[![GoDoc](https://godoc.org/github.com/claygod/coffer?status.svg)](https://godoc.org/github.com/claygod/coffer) [![Travis CI](https://travis-ci.org/claygod/coffer.svg?branch=master)](https://travis-ci.org/claygod/coffer) [![Go Report Card](https://goreportcard.com/badge/github.com/claygod/coffer)](https://goreportcard.com/report/github.com/claygod/coffer) [![codecov](https://codecov.io/gh/claygod/coffer/branch/master/graph/badge.svg)](https://codecov.io/gh/claygod/coffer)

# Coffer

Simple ACID* key-value database.

*is a set of properties of database transactions intended to guarantee validity even in the event of errors, power failures, etc.

Properties:
- Great throughput
- Admissible latency
- High reliability

ACID:
- Good durabilty
- Obligatory isolation
- Atomic operations
- Consistent transactions

## Usage

```golang
package main

import (
	"fmt"

	"github.com/claygod/coffer"
	"github.com/claygod/coffer/domain"
	"github.com/claygod/coffer/examples"
)

const curDir = "./"

func main() {

	// STEP init
	hdlExch := domain.Handler(HandlerExchange)
	db, err, wrn := coffer.Db(curDir).Create()
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

	// STEP write
	if rep := db.Write("foo", []byte("bar")); rep.IsCodeWarning() {
		fmt.Sprintf("Write error: code `%d` msg `%s`", rep.Code, rep.Error)
		return
	}

	// STEP read
	if rep := db.Read("foo"); rep.IsCodeWarning() {
		fmt.Sprintf("Read error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Println(string(rep.Data))

}
```
## How data is stored

## Data loading after an incorrect shutdown

## Benchmark

- BenchmarkCofferWriteParallel32LowConcurent-4		100000	12933 ns/op
- BenchmarkCofferTransactionSequence-4			2000		227928 ns/op
- BenchmarkCofferTransactionPar32NotConcurent-4	100000	4132 ns/op
- BenchmarkCofferTransactionPar32HalfConcurent-4	100000	4199 ns/op

### Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>
