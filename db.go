package coffer

// Coffer
// Db configurator
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

/*
Configurator - generates a configuration, and based on this configuration creates a database
*/
type Configurator struct {
	config   *Config
	handlers domain.HandlersRepository
}

/*
Db - specify the working directory in which the database will store its files. For a new database
the directory should be free of files with the extensions log, check, checkpoint.
*/
func Db(dirPath string) *Configurator {
	jCnf := &journal.Config{
		BatchSize:              defaultBatchSize,
		LimitRecordsPerLogfile: defaultLimitRecordsPerLogfile,
	}
	ucCnf := &usecases.Config{
		FollowPause:             defaultFollowPause,
		LogsByCheckpoint:        defaultLogsByCheckpoint, // after processing every N log files, the follower makes a new checkpoint
		DirPath:                 dirPath,
		AllowStartupErrLoadLogs: defaultAllowStartupErrLoadLogs, // if errors were detected during loading, is it possible to continue (by default it is possible)
		MaxKeyLength:            defaultMaxKeyLength,
		MaxValueLength:          defaultMaxValueLength,
		RemoveUnlessLogs:        defaultRemoveUnlessLogs, // clean logs after use
	}
	rcCnf := &resources.Config{
		LimitMemory: defaultLimitMemory, // minimum available memory (bytes)
		LimitDisk:   defaultLimitDisk,   // minimum free disk space
		DirPath:     dirPath,            //
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

/*
Create - creating a database. This operation should be the last in the configuration chain.
*/
func (c *Configurator) Create() (*Coffer, error, error) {
	return new(c.config, c.handlers)
}

/*
Handler - add a handler to the configuration. If a handler with such a key exists, it will be overwritten.
*/
func (c *Configurator) Handler(key string, value *domain.Handler) *Configurator {
	c.handlers.Set(key, value)
	return c
}

/*
Handlers - add several handlers to the configuration. Duplicate handlers will be overwritten.
*/
func (c *Configurator) Handlers(hdls map[string]*domain.Handler) *Configurator {
	for key, value := range hdls {
		c.handlers.Set(key, value)
	}
	return c
}

/*
BatchSize - the maximum number of records that a batch inside a database can add at a time
(this applies to setting up internal processes, this does not apply to the number of records added at a time).
Decreasing this parameter slightly improves the `latency` (but not too much). Increasing this parameter
slightly degrades the `latency`, but at the same time increases the throughput `throughput`.
*/
func (c *Configurator) BatchSize(value int) *Configurator {
	c.config.JournalConfig.BatchSize = value
	return c
}

/*
LimitRecordsPerLogfile - the number of operations to be written to one log file.
A small number will make the database very often create
new files, which will adversely affect the speed of the database.
A large number reduces the number of pauses for creation
files, but the files become larger.
*/
func (c *Configurator) LimitRecordsPerLogfile(value int) *Configurator {
	c.config.JournalConfig.LimitRecordsPerLogfile = int64(value)
	return c
}

/*
FollowPause - the size of the time interval for starting the `Follow` interactor,
which analyzes old logs and periodically creates new checkpoints.
*/
func (c *Configurator) FollowPause(value time.Duration) *Configurator {
	c.config.UsecasesConfig.FollowPause = value
	return c
}

/*
LogsByCheckpoint - after how many completed log files it is necessary to create a new checkpoint (the smaller
the number, the more often we create). For good performance, it’s better not to do it too often.
*/
func (c *Configurator) LogsByCheckpoint(value int) *Configurator {
	c.config.UsecasesConfig.LogsByCheckpoint = int64(value)
	return c
}

/*
AllowStartupErrLoadLogs - the option allows the database to work at startup,
even if the last log file was completed incorrectly, i.e. the last record is corrupted
(a typical situation for an abnormal shutdown). By default, the option is enabled.
*/
func (c *Configurator) AllowStartupErrLoadLogs(value bool) *Configurator {
	c.config.UsecasesConfig.AllowStartupErrLoadLogs = value
	return c
}

/*
MaxKeyLength - the maximum allowed key length.
*/
func (c *Configurator) MaxKeyLength(value int) *Configurator {
	c.config.UsecasesConfig.MaxKeyLength = value
	return c
}

/*
MaxValueLength - the maximum size of the value to write.
*/
func (c *Configurator) MaxValueLength(value int) *Configurator {
	c.config.UsecasesConfig.MaxValueLength = value
	return c
}

/*
RemoveUnlessLogs - option to delete old files. After `Follow` created a new checkpoint,
with the permission of this option, it now removes the unnecessary operation logs.
If for some reason you need to store the entire log of operations, you can disable this option,
but be prepared for the fact that this will increase the consumption of disk space.
*/
func (c *Configurator) RemoveUnlessLogs(value bool) *Configurator {
	c.config.UsecasesConfig.RemoveUnlessLogs = value
	return c
}

/*
LimitMemory - the minimum size of free RAM at which the database stops
performing operations and stops to avoid data loss.
*/
func (c *Configurator) LimitMemory(value int) *Configurator {
	c.config.ResourcesConfig.LimitMemory = int64(value)
	return c
}

/*
LimitDisk - the minimum amount of free space on the hard drive at which
the database stops performing operations and stops to avoid data loss.
*/
func (c *Configurator) LimitDisk(value int) *Configurator {
	c.config.ResourcesConfig.LimitDisk = int64(value)
	return c
}

/*
MaxRecsPerOperation - the maximum number of records that can be involved in one operation.
*/
func (c *Configurator) MaxRecsPerOperation(value int) *Configurator {
	c.config.MaxRecsPerOperation = value
	return c
}
