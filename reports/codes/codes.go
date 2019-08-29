package codes

// Coffer
// Config
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

type Code int64

const (
	//TODO: коды в отдельный пакет
	Ok Code = iota // выполнено без замечаний

	Warning // выполнено но с замечаниями

	Error                    // не выполнено, но работать дальше можно
	ErrRecordLimitExceeded   // превышен лимит записей на одну операцию
	ErrExceedingMaxKeyLength // слишком длинный ключ
	ErrHandlerNotFound       // не найден хэндлер
	ErrParseRequest          //  не получилось подготовить запрос для логгирования
	ErrResources             // не хватает ресурсов
	ErrReadRecords           //ошибка считывания записей для транзакции (при отсутствии хоть одной записи транзакцию нельзя проводить)
	ErrHandlerReturn         //найденный и загруженный хандлер вернул ошибку
	ErrHandlerResponse       // хандлер вернул неполные ответы

	Panic // не выполнено, дальнейшая работа с БД невозможна
	PanicStopped
	PanicWAL
)
