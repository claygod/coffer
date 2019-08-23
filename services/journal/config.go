package journal

// Coffer
// Journal (config)
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

// import (
// 	"log"
// 	"os"
// 	"runtime/pprof"
// 	"testing"

// 	//"time"

// 	"github.com/claygod/coffer/services/filenamer"
// 	"github.com/claygod/tools/batcher"
// )

type Config struct {
	//DirPath                string
	BatchSize              int
	LimitRecordsPerLogfile int64
}

const (
	stateStopped int64 = iota
	stateStarted
	statePanic
)
