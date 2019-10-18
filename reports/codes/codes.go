package codes

// Coffer
// Config
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

/*
Code - response code type
*/
type Code int64

const (
	//Ok - done without errors
	Ok Code = iota // выполнено без замечаний
	//Error - completed with errors, but you can continue to work
	Error // не выполнено, но работать дальше можно
	//ErrRecordLimitExceeded - record limit per operation exceeded
	ErrRecordLimitExceeded // превышен лимит записей на одну операцию
	//ErrExceedingMaxValueSize - value too long
	ErrExceedingMaxValueSize // слишком длинное значение
	//ErrExceedingMaxKeyLength - key is too long
	ErrExceedingMaxKeyLength // слишком длинный ключ
	//ErrExceedingZeroKeyLength - key too short
	ErrExceedingZeroKeyLength // слишком короткий ключ
	//ErrHandlerNotFound - not found handler
	ErrHandlerNotFound // не найден хэндлер
	//ErrParseRequest - failed to prepare a request for logging
	ErrParseRequest // не получилось подготовить запрос для логгирования
	//ErrResources - not enough resources
	ErrResources // не хватает ресурсов
	//ErrNotFound - no keys found
	ErrNotFound // не найдены ключи
	//ErrReadRecords - error reading records for a transaction (in the absence of at least one record, a transaction cannot be performed)
	ErrReadRecords // ошибка считывания записей для транзакции (при отсутствии хоть одной записи транзакцию нельзя проводить)
	//ErrHandlerReturn - found and loaded handler returned an error
	ErrHandlerReturn // найденный и загруженный хандлер вернул ошибку
	//ErrHandlerResponse - handler returned incomplete answers
	ErrHandlerResponse // хандлер вернул неполные ответы
	//Panic - not completed, further work with the database is impossible
	Panic // не выполнено, дальнейшая работа с БД невозможна
	//PanicStopped - the database is stopped, so you can’t work with it
	PanicStopped
	//PanicWAL - operation logging error, database stopped
	PanicWAL
)
