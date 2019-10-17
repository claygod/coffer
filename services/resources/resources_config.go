package resources

// Coffer
// Resources Config
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"time"
)

/*
Config - for ResourcesControl.
*/
type Config struct {
	LimitMemory int64 // minimum available memory (bytes)
	LimitDisk   int64 // minimum free disk space
	DirPath     string
}

const timeRefresh time.Duration = 1 * time.Millisecond
