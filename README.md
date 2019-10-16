[![GoDoc](https://godoc.org/github.com/claygod/coffer?status.svg)](https://godoc.org/github.com/claygod/coffer) [![Travis CI](https://travis-ci.org/claygod/coffer.svg?branch=master)](https://travis-ci.org/claygod/coffer) [![Go Report Card](https://goreportcard.com/badge/github.com/claygod/coffer)](https://goreportcard.com/report/github.com/claygod/coffer) [![codecov](https://codecov.io/gh/claygod/coffer/branch/master/graph/badge.svg)](https://codecov.io/gh/claygod/coffer)

# Coffer

Simple ACID* key-value database. At medium or even low `latency` provides
a large `throughput` without sacrificing ACID database properties.
The database makes it possible to create record headers at its discretion
and use them as transactions.

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

## Table of Contents

 * [Usage](#Usage)
 * [Examples](#Examples)
 * [API](#api)
      + [Methods](#methods)
 * [Config](#config)
      + [Handler](#Handler)
	      - [Handler example without using argument](#Handler-example-without-using-argument)
	      - [An example of a handler using an argument](#An-example-of-a-handler-using-an-argument)
 * [Launch](#Launch)
      - [Start](#Start)
      - [Follow](#Follow)
 * [Data storage](#Data-storage)
      - [Data loading after an incorrect shutdown](#Data-loading-after-an-incorrect-shutdown)
 * [Error codes](#Error-codes)
      - [Code List](#Code-List)
      - [Checking codes through methods](#Checking-codes-through-methods)
 * [Benchmark](#Benchmark)
 * [Dependencies](#Dependencies)
 * [ToDo](#TODO)

## Usage

```golang
package main

import (
	"fmt"

	"github.com/claygod/coffer"
	"github.com/claygod/coffer/domain"
)

const curDir = "./"

func main() {

	// STEP init
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

### Examples

You can find many examples of the use of transactions on these paths.

- `Quick start` https://github.com/claygod/coffer/tree/master/examples/quick_start
- `Finance` https://github.com/claygod/coffer/tree/master/examples/finance

## API

After the operation is completed, the database returns a report containing:
- code (error codes here: `github.com/claygod/coffer/reports/codes`)
- error
- data
- other details
Reporting structures here: `github.com/claygod/coffer/reports`

### Methods

* Start
* Stop
* StopHard
* Save
* Write
* WriteList
* WriteListUnsafe
* Read
* ReadList
* ReadListUnsafe
* Delete
* DeleteListStrict
* DeleteListOptional
* Transaction
* Count
* CountUnsafe
* RecordsList
* RecordsListUnsafe
* RecordsListWithPrefix
* RecordsListWithSuffix

Attention! All requests whose names contain `Unsafe` can be executed as if running,
and when the database is stopped (not running). In the second case, you cannot make queries in parallel,
otherwise, database consistency will be compromised and data will be lost.
Other methods work only if the database is running.

#### Start

Запустить БД. Большинство методов работает только если БД запущена.
Пи запуске включается `Follow` интерактор, который следит за актуальностью текущего чекпоинта.

#### Stop

Stop db. If you want to periodically stop and start the database in your application,
you may want to create a new client after the stop.

#### Write

Write a new record in the database, specifying the key and value.
Their length must satisfy the requirements specified in the configuration.

#### WriteList

Write several records to the database by specifying `map` in the arguments.
Important: this argument is a reference; it cannot be changed in the calling code!

#### WriteListUnsafe

Write several records to the database by specifying `map` in the arguments.
This method exists in order to fill it up a little faster before starting the database.
The method does not imply concurrent use.

#### Read

Read one entry from the database. In the received `report` there will be a result code, and if it is positive,
that will be the value in the `data` field.

#### ReadList

Read a few entries. There is a limit on the maximum number of readable entries in the configuration.
In addition to the found records, a list of not found records is returned.

#### ReadListUnsafe

Read a few entries. The method can be called when the database is stopped (not running).
The method does not imply concurrent use.

#### Delete

Remove a single record.

#### DeleteListStrict

Delete several records, but only if they are all in the database. If at least one entry is missing,
then no record will be deleted.

#### DeleteListOptional

Delete multiple entries. Those entries from the list that will be found in the database will be deleted.

#### Transaction

Execute a transaction. The transaction handler must be registered in the database at the stage
of creating and configuring the database. Responsibility for the consistency of the functionality
of transaction handlers between different database launches rests with the database user.
The transaction returns the new values stored in the database.

#### Count

Get the number of records in the database. A query can only be made to a running database

#### CountUnsafe

Get the number of records in the database. Queries to a stopped / not running database cannot be done in parallel!

#### RecordsList

Get a list of all database keys. With a large number of records in the database, the query will be slow, so use
its only in case of emergency. The method only works when the database is running.

#### RecordsListUnsafe

Get a list of all database keys. With a large number of records in the database, the query will be slow, so use
its only in case of emergency. When using a query with a stopped / not running database, competitiveness
prohibited.

#### RecordsListWithPrefix

Get a list of all the keys having prefix specified in the argument (start with that string).

#### RecordsListWithSuffix

Get a list of all the keys that have the specified argument suffix (ending).

## Config

It is enough to indicate the path to the database directory, and all configuration parameters will be set to default:

	cof, err, wrn := Db(dirPath) . Create()

Дефолтные значения можно увидеть в файле `/config.go` .
Однако каждый из параметров можно сконфигурировать:

```golang
	Db(dirPath).
	BatchSize(batchSize).
	LimitRecordsPerLogfile(limitRecordsPerLogfile).
	FollowPause(100*time.Second).
	LogsByCheckpoint(1000).
	AllowStartupErrLoadLogs(true).
	MaxKeyLength(maxKeyLength).
	MaxValueLength(maxValueLength).
	MaxRecsPerOperation(1000000).
	RemoveUnlessLogs(true).
	LimitMemory(100 * 1000000).
	LimitDisk(1000 * 1000000).
	Handler("handler1", &handler1).
	Handler("handler2", &handler2).
	Handlers(map[string]*handler).
	Create()
```
	
### Db

Specify the working directory in which the database will store its files. For a new database
the directory should be free of files with the extensions log, check, checkpoint.

### BatchSize

The maximum number of records that a batch inside a database can add at a time (this applies to setting up internal processes, this does not apply to the number of records added at a time). Decreasing this parameter slightly improves the `latency` (but not too much). Increasing this parameter slightly degrades the `latency`, but at the same time increases the throughput `throughput`.

### LimitRecordsPerLogfile

The number of operations to be written to one log file. A small number will make the database very often create
new files, which will adversely affect the speed of the database. A large number reduces the number of pauses for creation
files, but the files become larger.

### FollowPause

The size of the time interval for starting the `Follow` interactor, which analyzes old logs and periodically creates
new checkpoints.

### LogsByCheckpoint

After how many completed log files it is necessary to create a new checkpoint (the smaller the number, the more often we create).
For good performance, it’s better not to do it too often.

### AllowStartupErrLoadLogs

The option allows the database to work at startup, even if the last log file was completed incorrectly, i.e. the last record is corrupted (a typical situation for an abnormal shutdown). By default, the option is enabled.

### MaxKeyLength

The maximum allowed key length.

### MaxValueLength

The maximum size of the value to write.

### MaxRecsPerOperation

The maximum number of records that can be involved in one operation.

### RemoveUnlessLogs

Option to delete old files. After `Follow` created a new checkpoint, with the permission of this option,
it now removes the unnecessary operation logs. If for some reason you need to store the entire log of operations,
you can disable this option, but be prepared for the fact that this will increase the consumption of disk space.

### LimitMemory

The minimum size of free RAM at which the database stops performing operations and stops to avoid data loss.

### LimitDisk

The minimum amount of free space on the hard drive at which the database stops performing operations and stops to avoid data loss.

### Handler

Add transaction handler. It is important that for different launches of the same database, the name of the handler
and the results of its work are idempotent. Otherwise, at different times, with different starts, handlers will
work differently, which will lead to a violation of data consistency.
If you intend to make changes to handlers over time, adding a version number to the key may help streamline this process.

Conditions:
- The argument passed to the handler must be a number, a slice of bytes.
- If you need to transfer complex structures, they need to be serialized into bytes.
- The handler can only operate on existing records.
- The handler cannot delete database records.
- The handler at the end of the work should return the new values of all the requested records.
- The number of entries modified by the header is set in the `MaxRecsPerOperation` configuration

#### Handler example without using argument

```golang
func HandlerExchange(arg []byte, recs map[string][]byte) (map[string][]byte, error) {
	if arg != nil {
		return nil, fmt.Errorf("Args not null.")
	} else if len(recs) != 2 {
		return nil, fmt.Errorf("Want 2 records, have %d", len(recs))
	}
	recsKeys := make([]string, 0, 2)
	recsValues := make([][]byte, 0, 2)
	for k, v := range recs {
		recsKeys = append(recsKeys, k)
		recsValues = append(recsValues, v)
	}
	out := make(map[string][]byte, 2)
	out[recsKeys[0]] = recsValues[1]
	out[recsKeys[1]] = recsValues[0]
	return out, nil
}
```

#### An example of a handler using an argument

```golang
func HandlerDebit(arg []byte, recs map[string][]byte) (map[string][]byte, error) {
	if arg == nil || len(arg) != 8 {
		return nil, fmt.Errorf("Invalid Argument: %v.", arg)
	} else if len(recs) != 1 {
		return nil, fmt.Errorf("Want 1 record, have %d", len(recs))
	}
	delta := bytesToUint64(arg)
	var recKey string
	var recValue []byte
	for k, v := range recs {
		recKey = k
		recValue = v
	}
	if len(recValue) != 8 {
		return nil, fmt.Errorf("The length of the value in the record is %d bytes, but 8 bytes are needed", len(recValue))
	}
	curAmount := bytesToUint64(recValue)
	newAmount := curAmount + delta
	if curAmount > newAmount {
		return nil, fmt.Errorf("Account overflow. There is %d, a debit of %d.", curAmount, delta)
	}
	return map[string][]byte{recKey: uint64ToBytes(newAmount)}, nil
}
```

### Handlers

Add multiple handlers to the database at a time. Important: handlers with matching keys are overwritten.

### Create

The required command (must be the last one) finishes the configuration and creates the database.

## Launch

### Start

At start, the last number should be a checkpoint. If this is not so, then the stop was incorrect.
Then the last uncorrected checkpoint and all the logs after it are loaded until it is possible. On a beat log
or the last log, the download ends. The database creates a new checkpoint, and after that the answer is returned,
and the database is ready to start.

### Follow

After the database is launched, it writes all operations to the log. As a result, the log can grow very much.
If in the end, at the end of the application, the database is correctly stopped, a new checkpoint will appear,
and at the next start, the data will be taken from it.
However, the stop may not be correct, and a new checkpoint will not be created.

In this case, at a new start, the database will be forced to load the old checkpoint, and re-perform all operations
that were completed and recorded in the log. This can turn out to be quite significant in time, and as a result,
the database will take longer to load, which is not always acceptable for applications.

That is why there is a follower mechanism in the database that methodically goes through the logs in the process of
the database and periodically creates checkpoints that are much closer to the current moment.
Also, the follower has the function of cleaning old logs and checkpoints to free up space on your hard drive.

## Data storage

Your data is stored as files in the directory that you specified when creating the database.
Files with the extension `log` contain a description of the operations performed.
Files with the extension `checkpoint` contain snapshots of the state of the database at a certain point.
Files with the `check` extension contain an incomplete snapshot of the state of the database.
Using the `RemoveUnlessLogs` configuration parameter, you can order the database to delete old
and unnecessary files in order to save disk space.

If the database is stopped in the normal mode, then the last file written to the disk will be the file `checkpoint`,
and its number will be the maximum. If the database is stopped incorrectly, then most likely the file with
the extension `log` or` check` will have the maximum number.

Attention! until the database is completely stopped, it is forbidden to carry out any operations with database files.

If you want to copy the database somewhere, you must copy the entire contents of the directory.
If you want to take a minimum of files when copying, then you need to copy the file with the `checkpoint` extension,
which has the maximum number, and all files with the` log` extension, which have a number larger than the copied checkpoint file.

### Data loading after an incorrect shutdown

If the application using the database is not completed correctly, then at the next boot, the database will try
to find the last valid snapshot of the `checkpoint` state. Having found this file, the database will load it,
after which it will load all the `log` files with large numbers. We expect that the last `log` file may
not be completely filled, because during the recording work could be interrupted. Therefore, the download
from the damaged file will be performed to the damaged (unrecorded) section, after which the database download
is considered complete. At the end of the download, the database creates a new `checkpoint`.
If system crashes occur during the start (load) of the database, errors and violation of data consistency are possible.

## Error Codes

Error codes are stored here: `github.com/claygod/coffer/reports/codes`
If the `Ok` code is received, then the operation is complete. If the Code contains `Error`, then the operation
has not been completed, it is incomplete, or completed with an error, but you can continue working with the database.
If the code contains `Panic`, then the state of the database is such that you cannot continue to work with it.

### Code List

- Ok - done without comment
- Error - not completed or not fully completed, but you can continue to work
- ErrRecordLimitExceeded - record limit per operation exceeded
- ErrExceedingMaxValueSize - value is too long
- ErrExceedingMaxKeyLength - key is too long
- ErrExceedingZeroKeyLength - key is too short
- ErrHandlerNotFound - no handler found
- ErrParseRequest - failed to prepare a request for logging
- ErrResources - not enough resources
- ErrNotFound - no keys found
- ErrReadRecords - error reading records for a transaction (in the absence of at least one record, a transaction cannot be performed)
- ErrHandlerReturn - the found and downloaded handler returned an error
- ErrHandlerResponse - handler returned incomplete answers
- Panic - not done, further work with the database is impossible
- PanicStopped - application stopped
- PanicWAL - an error in the operation log

### Checking codes through methods

In order not to export to an application that works with a database, reports have methods:

- IsCodeOk - done without comment
- IsCodeError - not completed or not fully completed, but you can continue to work
- IsCodeErrRecordLimitExceeded - record limit for one operation is exceeded
- IsCodeErrExceedingMaxValueSize - value is too long
- IsCodeErrExceedingMaxKeyLength - key is too long
- IsCodeErrExceedingZeroKeyLength - key is too short
- IsCodeErrHandlerNotFound - no handler found
- IsCodeErrParseRequest - failed to prepare a request for logging
- IsCodeErrResources - not enough resources
- IsCodeErrNotFound - no keys found
- IsCodeErrReadRecords - error reading records for a transaction (in the absence of at least one record, a transaction cannot be performed)
- IsCodeErrHandlerReturn - the found and loaded handler returned an error
- IsCodeErrHandlerResponse - handler returned incomplete answers
- IsCodePanic - not completed, further work with the database is impossible
- IsCodePanicStopped - application stopped
- IsCodePanicWAL - error in the operation log

It is not very convenient to make large switches to check the received codes. You can limit yourself to just three checks:

- IsCodeOk - done without comment
- IsCodeError - not completed or not fully completed, but you can continue to work (covers ALL errors)
- IsCodePanic - not completed, further work with the database is not possible (covers ALL panics)

## Benchmark

- BenchmarkCofferWriteParallel32LowConcurent-4		100000	12933 ns/op
- BenchmarkCofferTransactionSequence-4			2000		227928 ns/op
- BenchmarkCofferTransactionPar32NotConcurent-4	100000	4132 ns/op
- BenchmarkCofferTransactionPar32HalfConcurent-4	100000	4199 ns/op

## Dependencies

- github.com/shirou/gopsutil/disk
- github.com/shirou/gopsutil/mem
- github.com/sirupsen/logrus

## TODO

- [x] the log should start a new log at startup
- [x] deal with the names of checkpoints and logs (numbering logic)
- [x] launch and follower operation
- [x] cleaning unwanted logs with a follower
- [ ] provide an opportunity not to delete old logs, add a test!
- [x] loading from broken files to stop loading, but work continued (AllowStartupErrLoadLogs)
- [x] cyclic loading of checkpoints until they run out (with errors)
- [x] return not of errors, but of progress reports
- [x] add DeleteOptional, including in Operations
- [x] test Count
- [x] Write test
- [x] Read test
- [x] Delete test
- [x] Transaction test
- [x] test RecordsList
- [x] test RecordsListUnsafe
- [x] test RecordsListWithPrefix
- [x] test RecordsListWithSuffix
- [x] ReadListUnsafe test
- [x] boot test with a broken log (last, the rest are ok)
- [x] download test with broken checkpoint
- [x] boot test with a broken log and another log following it
- [x] transaction usage test
- [x] for convenience of testing do WriteUnsafe
- [x] ~~ what is WriteUnsafeRecord in Checkpoint for? (for recording at startup?) ~~ alternative to WriteListUnsafe (faster)
- [x] benchmark entries competitive and non-competitive
- [x] benchmark reading competitive
- [ ] benchmark write and read in competitive mode
- [x] parallel competitive benchmark
- [ ] at boot - when broken files wrn may return, not err
- [x] deal with the log and the batch, why do records go to the next log when recording quickly
- [x] interception of panics at the root of the application and at the level of usecases
- [ ] ~~ during transactions, you can delete some of the entries from participating (! need for a question!) ~~
- [x] testing helper helpers
- [x] when creating a database, immediately add a list of handlers, because and loading from the logs also happens immediately
- [x] add a convenient configurator when creating a database
- [ ] translate comments into English
- [ ] clear code from old artifacts
- [ ] create a directory for documentation
- [x] create a directory for examples
- [x] make a simple example with writing, transaction and reading
- [x] make an example with financial transactions
- [ ] error handling example
- [ ] banish the linter and eliminate all incorrectness in the code
- [x] add Usage / Quick start text to readme
- [x] description of error codes
- [x] configuration description
- [x] in the description specify third-party packages (as dependencies)
- [x] reports add methods for checking for all errors in the spirit of IsErrBlahBlahBlah
- [x] transfer all imported packages to distribution
- [x] switch the use of WriteUnsafeRecord to WriteListUnsafe
- [x] addReadListUnsafe for readability when the database is stopped
- [x] add RecordsListUnsafe, which can work with the database stopped and running
- [x] obtaining a list of keys with a condition by the prefix RecordsListWithPrefix
- [x] obtaining a list of keys with a condition by the suffix RecordsListWithSuffix
- [x] remove the Save method
- [x] when a transaction returns new values ​​in the report
- [x] in tests check return value
- [x] start numbering with large digits, say with a million, or a billion (more convenient for sorting files)
- [ ] all public methods give a correct description-comment
- [ ] return error and warning in Create method


### Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>
