[![GoDoc](https://godoc.org/github.com/claygod/coffer?status.svg)](https://godoc.org/github.com/claygod/coffer) [![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go) [![Travis CI](https://travis-ci.org/claygod/coffer.svg?branch=master)](https://travis-ci.org/claygod/coffer) [![Go Report Card](https://goreportcard.com/badge/github.com/claygod/coffer)](https://goreportcard.com/report/github.com/claygod/coffer) [![codecov](https://codecov.io/gh/claygod/coffer/branch/master/graph/badge.svg)](https://codecov.io/gh/claygod/coffer)


# Coffer

Simply ACID* key-value database. At the medium or even low `latency` it tries to
provide greater `throughput` without losing the ACID properties of the database. The
database provides the ability to create record headers at own discretion and use them
as transactions. The maximum size of stored data is limited by the size of the
computer's RAM.

*is a set of properties of database transactions intended to guarantee validity even in
the event of errors, power failures, etc.

Properties:
- high throughput
- tolerated latency
- high reliability

ACID:
- good durabilty
- compulsory isolation
- atomic operations
- consistent transactions

## Table of Contents

 * [Usage](#Usage)
 * [Examples](#Examples)
 * [API](#api)
      + [Methods](#methods)
 * [Config](#config)
      + [Handler](#Handler)
	      - [Example of a handler without using an argument](#Example-of-a-handler-without-using-an-argument)
	      - [Example of a handler using an argument](#Example-of-a-handler-using-an-argument)
 * [Launch](#Launch)
      - [Start](#Start)
      - [Follow](#Follow)
 * [Data storage](#Data-storage)
      - [Data loading after an incorrect shutdown](#Data-loading-after-an-incorrect-shutdown )
 * [Error codes](#Error-codes)
      - [Code List](#Code-List)
      - [Code checks through methods](#Code-checks-through-methods)
 * [Benchmark](#Benchmark)
 * [Dependencies](#Dependencies)
 * [ToDo](#TODO)

## Usage

```golang
package main

import (
	"fmt"

	"github.com/claygod/coffer"
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
	if rep := db.Write("foo", []byte("bar")); rep.IsCodeError() {
		fmt.Sprintf("Write error: code `%d` msg `%s`", rep.Code, rep.Error)
		return
	}

	// STEP read
	rep := db.Read("foo")
	rep.IsCodeError()
	if rep.IsCodeError() {
		fmt.Sprintf("Read error: code `%v` msg `%v`", rep.Code, rep.Error)
		return
	}
	fmt.Println(string(rep.Data))
}
```

### Examples

Use the following links to find many examples how use the transactions:

- `Quick start` https://github.com/claygod/coffer/tree/master/examples/quick_start
- `Finance` https://github.com/claygod/coffer/tree/master/examples/finance

## API

Started DB returns reports after has performed an operation. Reports containing:
- code (error codes here: `github.com/claygod/coffer/reports/codes`)
- error
- data
- other details
Reporting structures read here:: `github.com/claygod/coffer/reports`

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

Pay attention!

All requests which names contain `Unsafe` can be usually executed in both cases: when the
database is running or stopped (not running). In the second case (when DB is stopped),
should not make requests in parallel, because in this case the consistency of DB can
be compromised and data lost.

Other methods work only if the database is running.

#### Start

The `Follow` interactor turns on while running a database. It controls the relevance of the
current checkpoint.

#### Stop

Stop DB. If you want to periodically stop and start the database in your application, probably,
you may want to create a new client when the DB has been stopped.

#### Write

Write a new record in a database specifying the key and value. Their length must satisfy the
requirements specified in configuration files.

#### WriteList

Write several records to the database specifying corresponding `map` in the arguments.

Strict mode (strictMode=true):
	The operation performs if there are no records with these keys.
	A list of existed records is returned.
	
Optional mode (strictMode=false):
	The operation performs regardless of whether there are records with these keys or not.
	A list of existed records is returned.
	
Important: this argument is a reference argument; it cannot be changed in the called code!

#### WriteListUnsafe

Write several records to the database specified the corresponding map in the arguments.
This method exists in order to fill the database faster before it starts.
The method is not for parallel use.

#### Read

Read one record from the database. In the received report there will be a result code.
If it is positive, that means that the value in the right data field.

#### ReadList

Read several records. There is a limit on the maximum number of readable records in the
configuration. Except found records the list of not found records is returned.

#### ReadListUnsafe

Read several records. The method can be called when the database is stopped (not running).
The method is not for parallel use.

#### Delete

Remove a single record.

#### DeleteList

Strict mode (true):
Delete several records. It is possible only if all records are in the database.
If at least there is a lack of one record, none of records will be deleted.

Optional mode (false):
Delete several records. All found records from the list in will be deleted in DB.

#### Transaction

Make a transaction. The transaction should be added in the database at the stage of creating
and configuring. The user of the database is responsible for the consistency of the functionality of
transaction handlers between different runs of the database.
The transaction returns new values which are stored in the DB.

#### Count

Get the number of records in the database. A request can be made only when the database has started.

#### CountUnsafe

Get the number of records in the database. Requests to a stopped (or not running),
database cannot be made in parallel!

#### RecordsList

Get a list of all database keys. With a large number of records in the database, the request
will be slow. Use it only at great need to avoid problems.
The method works only when the database is running.

#### RecordsListUnsafe

Get a list of all database keys. With a large number of records in the database, the request
will be slow. Use it only at great need to avoid problems. The method is not for parallel use
while using a request when database is stopped (or not running).

#### RecordsListWithPrefix

Get a list of all keys with prefix specified in the argument (prefix is the begging of record string).

#### RecordsListWithSuffix

Get a list of all the keys with specified suffix in the argument (Suffix is in the ending of record string).

## Config

If you specify the path to the database directory all configuration parameters will be
reset to the default:

	cof, err, wrn := Db(dirPath) . Create()

Default values can be found in the `/config.go` file. But each of the parameters can be
configured:

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

Specify the work directory where the database will store files. In case of a new
database the directory should not contain files with the “log”, “check”, “checkpoint”
extensions.

### BatchSize

The maximum number of records which database can add at a time (this applies to
setting up internal processes; this does not apply to the number of records added at a
time).
Decreasing of this parameter slightly improves the `latency` (but not too much).
Increasing of this parameter slightly degrades the `latency`, but at the same time
increases the `throughput`.

### LimitRecordsPerLogfile

A number of operations which is going to be written to one log file. Small number
forces the database creates new files very often, and it adversely affects the speed of
the database. A big number reduces the number of pauses while creating files, but
the size of files increases.

### FollowPause

The size of the time interval for starting the `Follow` interactor, which analyzes old
logs and periodically creates new checkpoints.

### LogsByCheckpoint

The option specifies after how many full log files it is necessary to create a new
checkpoint (the smaller number, the more often it should be created). For good
productivity, it’s better not to do it too often.

### AllowStartupErrLoadLogs

The option allows the database works at startup, even if the last log file was
completed incorrectly, i.e. the last record is corrupted (a typical situation for an
abnormal shutdown). By default, the option is enabled.

### MaxKeyLength

This is the maximum allowable key length.

### MaxValueLength

This is the maximum size of the value length.

### MaxRecsPerOperation

This is the maximum number of records that is possible per operation.

### RemoveUnlessLogs

The option is for deleting old files. After `Follow` has created a new checkpoint, with the
permission of this option, it removes unnecessary operations logs. If for some reason
it’s needed to store the whole log of operations, this option can be disabled. But be
ready that this will increase the consumption of disk space.

### LimitMemory

This is the minimum size of free RAM. When this limit reaches, the database
terminates all operations and stops to avoid data loss.

### LimitDisk

This is the minimum amount of free space on the hard drive. When this limit reaches,
the database terminates all operations and stops to avoid data loss.

### Handler

Add a transaction handler. It is important that the name of the handler and the results
of its work should be idempotent while running the same database at different time.
Otherwise handlers will work differently and it will leads to a violation of data
consistency. If you intend to make changes to handlers time to time, adding a version
number to the key helps streamline this process.

Conditions:
- The argument passed to the handler must be a number (a slice of bytes).
- If you need to transfer complex structures, it must be serialized into bytes.
- The handler can only operate on existing records.
- The handler cannot delete database records.
- The handler should return the new values of all the requested records at the end of his work.
- The number of records modified with the header is specified in the `MaxRecsPerOperation` configuration.

#### Example of a handler without using an argument

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

#### Example of a handler using an argument

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

Add several handlers to the database at once. Important: handlers with matching keys
are overwritten.

### Create

A mandatory command (must be the last one) finishes the configuration and creates
the database.

## Launch

### Start

At starting DB, the number in the end should be a checkpoint. If it is not, the database
has been stopped incorrectly. In this “error” case the last available checkpoint and all
logs after the checkpoint are loaded until it is possible.

Load the data until it is possible and finish the loading (there is must be a broken log
(log with uncompleted data) or last log which has created before database has been
stopped incorrectly, so you can load all available data). After all available data has
loaded the database creates a new checkpoint. Only after it you can continue work
with code.

### Follow

After the database has been started, it writes all operations to a log. As a result, the
log file can greatly grow. If at the end of the application work the database is correctly
stopped, a new checkpoint appears. At the next start of DB, the data will be taken
from it.

But if the database is incorrectly stopped a new checkpoint will not be created. In this
case, at a new start of DB, the database loads the old checkpoint and re-performs all
operations that has been completed and recorded in the log. This process can take
much time, and as a result, the database will be loading for a long time (not always
acceptable for applications).

That is why there is the follower mechanism in the database that methodically goes
through the logs while working of the database and periodically creates checkpoints
which are closer to the current moment. Also, the follower has a functionality to clean
old logs and checkpoints in order to free up the space of hard drive.

## Data storage

Your data is stored as files in the directory that has been specified while creating the
database. Files with the log extension contain a description of completed operations.
Files with the `checkpoint` extension contain snapshots of the database state at a
certain point. Files with the `check` extension contain an incomplete snapshot of the
database state. Using the `RemoveUnlessLogs` configuration parameter, you can force
the database to delete old and unnecessary files in order to save the disk space.

If the database is stopped in the regular mode, the last file, which has been written to
the disk, is the `checkpoint` file. The number of the `checkpoint` will be the maximum
number. If the database is stopped incorrectly, most likely that files with the `log` or
`check` extensions will have the maximum number.

Attention! Before the database has not been completely stopped, it is forbidden to
carry out any operations with database files.

If you want to copy the database to somewhere, you must copy all content of the
directory. If you want to take a minimum of files while copying, then you need: to
copy the file with the `checkpoint` extension (which has the maximum number), and
all files with the `log` extension (which have bigger number than copied file with the
`checkpoint` extension).

### Data loading after an incorrect shutdown

If the application work, which is using the database, is not completed correctly, then
at the next staring the application, the database will try to find the last valid snapshot
of the `checkpoint` state.

When the file is found, the database will upload it, and then upload all the `log` files
with big numbers. We expect that the last `log` file might not be filled completely
because during the recording the work could be interrupted.

Only undamaged part is uploaded from the damaged file and after that the database
uploading is considering as completed.

At the end of the uploading, the database creates a new `checkpoint`. If system
crashes occur during the start (loading) of the database, it possible to get errors and
violation of data consistency.

## Error Codes

Error codes are stored here: `github.com/claygod/coffer/reports/codes`
If the `Ok` code is received, the operation is finished completely. If the Code contains
`Error` (the operation has not been completed or not fully completed, or completed
with an error), you can continue working with the database. If the code contains
`Panic`, you cannot continue working with the database because of it stage.

### Code List

- Ok - done without comment
- Error - not completed or not fully completed, but you can continue to work
- ErrRecordLimitExceeded - record limit per operation exceeded
- ErrExceedingMaxValueSize - value is too long
- ErrExceedingMaxKeyLength - key is too long
- ErrExceedingZeroKeyLength - key is too short
- ErrHandlerNotFound - no handler found
- ErrParseRequest – preparing of logging request was failed
- ErrResources - not enough resources
- ErrNotFound - no keys are found
- ErrReadRecords - reading records error for a transaction (if there is a lack of at least one record, a transaction cannot be performed)
- ErrHandlerReturn - found and uploaded handler returned an error
- ErrHandlerResponse - handler returned incomplete reply
- Panic - not finished, further work with the database is impossible
- PanicStopped – application has been stopped
- PanicWAL - an error occurred in the operation log

### Code checks through methods

In order not to export data to an application (which works with a database), reports
have methods:

- IsCodeOk - done without comment
- IsCodeError - not completed or not fully completed, but you can continue to work
- IsCodeErrRecordLimitExceeded - record limit for one operation is exceeded
- IsCodeErrExceedingMaxValueSize - value is too long
- IsCodeErrExceedingMaxKeyLength - key is too long
- IsCodeErrExceedingZeroKeyLength - key is too short
- IsCodeErrHandlerNotFound - no handler found
- IsCodeErrParseRequest - preparing of logging request was failed
- IsCodeErrResources - not enough resources
- IsCodeErrNotFound - no keys are found
- IsCodeErrReadRecords - reading records error for a transaction (if there is a lack of at least one record, a transaction cannot be performed)
- IsCodeErrHandlerReturn – found and uploaded handler returned an error
- IsCodeErrHandlerResponse - handler returned incomplete reply
- IsCodePanic - not finished, further work with the database is impossible
- IsCodePanicStopped – application has been stopped
- IsCodePanicWAL - error occurred in the operation log

It is not very convenient to make large switches to check the received codes. You can
limit yourself to just three checks:

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
- [x] study out the names of checkpoints and logs (numbering logic)
- [x] launch and work of follower
- [x] cleaning unnecessary logs via follower
- [ ] provide an opportunity not to delete old logs, add a test!
- [x] loading from broken files to stop loading, but work must continue (AllowStartupErrLoadLogs)
- [x] cyclic loading of checkpoints until they run out (with errors)
- [x] returns not errors, but reports of work that’s been completed
- [x] add DeleteOptional, and add it in Operations too
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
- [x] boot test with broken checkpoint
- [x] boot test with a broken log and another the log which follow after
- [x] transaction usage test
- [x] for convenience of testing do WriteUnsafe
- [x] ~~ what for WriteUnsafeRecord is need in Checkpoint ? (for recording at startup?) ~~ alternative to WriteListUnsafe (faster)
- [x] benchmark of competitive and non-competitive records
- [x] benchmark reading competitive
- [ ] benchmark write and read in competitive mode
- [x] benchmark competitive transactions in parallel mode
- [ ] at boot - when files are broken, the “wrn” may returns, not “err”
- [x] study out the log and the batch, why at fast record they get to the following log
- [x] interception of panics at the root of the application and at the level of use cases
- [ ] ~~ during transactions, you can delete some of the records from participating (! need for a question!) ~~ 
- [x] testing auxiliary helpers
- [x] while creating a database immediately add a list of handlers because the uploading from logs happens instantly
- [x] add a convenient configurator while creating a database
- [x] translate comments into English
- [x] clear code from old artifacts
- [ ] create a directory for documentation
- [x] create a directory for examples
- [x] make a simple example with writing, transaction and reading
- [x] make an example with financial transactions
- [ ] error handling example
- [ ] test the linter and eliminate all incorrectness in the code
- [x] add Usage / Quick start text to readme
- [x] description of error codes
- [x] configuration description
- [x] in the description specify third-party packages (as dependencies)
- [x] add methods for reports in order to check for all errors like IsErrBlahBlahBlah
- [x] transfer all imported packages to distribution
- [x] switch from WriteUnsafeRecord to WriteListUnsafe
- [x] add ReadListUnsafe for an ability to read when the database is stopped
- [x] add RecordsListUnsafe, which can work with the stopped and running database
- [x] get a list of keys with a condition of a prefix: RecordsListWithPrefix
- [x] get a list of keys with a condition of a suffix: RecordsListWithSuffix
- [x] remove the Save method
- [x] returns new values in the report during a transaction
- [x] check returned value in tests
- [x] start numbering with big numbers, for example million or a billion (more convenient for sorting files)
- [x] give a correct description-comment for all public methods
- [ ] create description for error returns and warnings in the Create method
- [ ] pause in the batcher - check its size, set the optimal size
- [x] add in the description that the data is stored both on disk and in memory during the operation of the database
- [ ] method for getting all log files and checkpoints
- [ ] method for viewing a log file
- [ ] method of viewing a checkpoint file
- [ ] the method of strict adding record into the database (only if the record with such a key has not already existed)

### Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>
