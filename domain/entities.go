package domain

// Coffer
// Domain entities
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

/*
Handler - обработчик в транзакции.
Правило написания обработчика простое: он может только изменять значения в
полученном массиве. Добавление новых ключей запрещено, т.к. фактически будет означать
добавление записей, а таковые уже могут быть. Уменьшение ключей будет означать удаление записей,
это возможно, т.к. все ключи залочены и одновременного доступа к ним не может быть.
*/
type Handler func(interface{}, map[string][]byte) (map[string][]byte, error)

type Operation struct {
	Code byte
	Body []byte
}

type Record struct {
	Key   string
	Value []byte
}
