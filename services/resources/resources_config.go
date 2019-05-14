package resources

// Coffer
// Resources Config
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"time"
)

type Config struct {
	LimitMemory int64 // minimum available memory (bytes)
	LimitDisk   int64 // minimum free disk space
	DickPath    string
}

const timeRefresh time.Duration = 200 * time.Millisecond
