package usecases

// Coffer
// Config
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"time"
)

type Config struct {
	FollowPause             time.Duration
	ChagesByCheckpoint      int64
	DirPath                 string
	AllowStartupErrLoadLogs bool
	RemoveUnlessLogs        bool // удаление логов после того, как они  попали в чекпоинт
	MaxKeyLength            int  //  = int(uint64(1)<<16) - 1
	MaxValueLength          int  //  = int(uint64(1)<<48) - 1
}

const (
	stateStopped int64 = iota
	stateStarted
	statePanic
)

const (
	extLog   string = ".log"
	extCheck string = ".check"
	extPoint string = "point"
	megabyte int64  = 1024 * 1024
)

const (
	codeWriteList byte = iota //codeWrite
	codeTransaction
	codeDeleteList
)

const (
	//TODO: коды в отдельный пакет
	CodeOk int64 = iota // выполнено без замечаний

	CodeWarning // выполнено но с замечаниями

	CodeError              // не выполнено, но работать дальше можно
	CodeErrHandlerNotFound // не найден хэндлер
	CodeErrParseRequest    //  не получилось подготовить запрос для логгирования
	CodeErrResources       // не хватает ресурсов
	CodeErrReadRecords     //ошибка считывания записей для транзакции (при отсутствии хоть одной записи транзакцию нельзя проводить)
	CodeErrHandlerReturn   //найденный и загруженный хандлер вернул ошибку
	CodeErrHandlerResponse // хандлер вернул неполные ответы

	CodePanic // не выполнено, дальнейшая работа с БД невозможна
	CodePanicStopped
	CodePanicWAL
)
