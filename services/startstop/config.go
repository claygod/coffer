package startstop

// Coffer
// StartStop (config)
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"time"
)

const (
	stateBlocked   int64 = -2
	stateReady     int64 = -1
	stateRun       int64 = 0
	maxIterations  int   = 1e10
	blockedBarrier int64 = 1e15
)

const pauseDefault time.Duration = 10 * time.Microsecond
