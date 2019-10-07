package coffer

// Coffer
// Db sucar
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"time"

	"github.com/claygod/coffer/domain"
	"github.com/claygod/coffer/services/journal"
	"github.com/claygod/coffer/services/repositories/handlers"
	"github.com/claygod/coffer/services/resources"
	"github.com/claygod/coffer/usecases"
)

type Configurator struct {
	config   *Config
	handlers domain.HandlersRepository
}

func Db(dirPath string) *Configurator {
	jCnf := &journal.Config{
		BatchSize:              defaultBatchSize,
		LimitRecordsPerLogfile: defaultLimitRecordsPerLogfile,
	}
	ucCnf := &usecases.Config{
		FollowPause:             defaultFollowPause,
		LogsByCheckpoint:        defaultLogsByCheckpoint, //после обработки каждых N файлов логов фолловер делает новый чекпоинт
		DirPath:                 dirPath,
		AllowStartupErrLoadLogs: defaultAllowStartupErrLoadLogs, // если при загрузке обнаружены ошибки, можно ли продолжать (по умолчанию можно)
		MaxKeyLength:            defaultMaxKeyLength,
		MaxValueLength:          defaultMaxValueLength,
		RemoveUnlessLogs:        defaultRemoveUnlessLogs, // чистим логи после использования
	}
	rcCnf := &resources.Config{
		LimitMemory: defaultLimitMemory, // minimum available memory (bytes)
		LimitDisk:   defaultLimitDisk,   // minimum free disk space
		DirPath:     dirPath,            // "/home/ed/goPath/src/github.com/claygod/coffer/test"
	}

	cnf := &Config{
		JournalConfig:       jCnf,
		UsecasesConfig:      ucCnf,
		ResourcesConfig:     rcCnf,
		MaxRecsPerOperation: defaultMaxRecsPerOperation,
		//MaxKeyLength:        100,
		//MaxValueLength:      10000,
	}

	hdls := handlers.New()

	db := &Configurator{
		config:   cnf,
		handlers: hdls,
	}
	return db
}

func (c *Configurator) Create() (*Coffer, error, error) {
	return new(c.config, c.handlers)
}

func (c *Configurator) Handler(key string, value *domain.Handler) *Configurator {
	c.handlers.Set(key, value)
	return c
}
func (c *Configurator) Handlers(hdls map[string]*domain.Handler) *Configurator {
	for key, value := range hdls {
		c.handlers.Set(key, value)
	}
	return c
}

func (c *Configurator) BatchSize(value int) *Configurator {
	c.config.JournalConfig.BatchSize = value
	return c
}
func (c *Configurator) LimitRecordsPerLogfile(value int) *Configurator {
	c.config.JournalConfig.LimitRecordsPerLogfile = int64(value)
	return c
}
func (c *Configurator) FollowPause(value time.Duration) *Configurator {
	c.config.UsecasesConfig.FollowPause = value
	return c
}
func (c *Configurator) LogsByCheckpoint(value int) *Configurator {
	c.config.UsecasesConfig.LogsByCheckpoint = int64(value)
	return c
}
func (c *Configurator) AllowStartupErrLoadLogs(value bool) *Configurator {
	c.config.UsecasesConfig.AllowStartupErrLoadLogs = value
	return c
}
func (c *Configurator) MaxKeyLength(value int) *Configurator {
	c.config.UsecasesConfig.MaxKeyLength = value
	return c
}
func (c *Configurator) MaxValueLength(value int) *Configurator {
	c.config.UsecasesConfig.MaxValueLength = value
	return c
}
func (c *Configurator) RemoveUnlessLogs(value bool) *Configurator {
	c.config.UsecasesConfig.RemoveUnlessLogs = value
	return c
}
func (c *Configurator) LimitMemory(value int) *Configurator {
	c.config.ResourcesConfig.LimitMemory = int64(value)
	return c
}
func (c *Configurator) LimitDisk(value int) *Configurator {
	c.config.ResourcesConfig.LimitDisk = int64(value)
	return c
}
func (c *Configurator) MaxRecsPerOperation(value int) *Configurator {
	c.config.MaxRecsPerOperation = value
	return c
}

// jCnf := &journal.Config{
// 	BatchSize:              batchSize,
// 	LimitRecordsPerLogfile: limitRecordsPerLogfile,
// }
// ucCnf := &usecases.Config{
// 	FollowPause:             100 * time.Second, //чтобы точно не включался
// 	LogsByCheckpoint:        1000,              //чтобы точно не включался
// 	DirPath:                 dirPath,           // "/home/ed/goPath/src/github.com/claygod/coffer/test",
// 	AllowStartupErrLoadLogs: true,
// 	MaxKeyLength:            maxKeyLength,
// 	MaxValueLength:          maxValueLength,
// 	RemoveUnlessLogs:        true, // чистим логи после использования
// }
// rcCnf := &resources.Config{
// 	LimitMemory: 1000 * megabyte, // minimum available memory (bytes)
// 	LimitDisk:   1000 * megabyte, // minimum free disk space
// 	DirPath:     dirPath,         // "/home/ed/goPath/src/github.com/claygod/coffer/test"
// }

// cnf := &Config{
// 	JournalConfig:       jCnf,
// 	UsecasesConfig:      ucCnf,
// 	ResourcesConfig:     rcCnf,
// 	MaxRecsPerOperation: 1000,
// 	//MaxKeyLength:        100,
// 	//MaxValueLength:      10000,
// }
