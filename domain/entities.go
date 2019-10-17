package domain

// Coffer
// Domain entities
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

/*
Handler - handler in a transaction.
The rule for writing a handler is simple: it can only change values in
the resulting array. Adding new keys is prohibited, because will actually mean
adding records, and those may already be. Decreasing the keys will mean deleting the entries,
it is possible because all keys are locked and simultaneous access to them cannot be.
*/
type Handler func([]byte, map[string][]byte) (map[string][]byte, error)

/*
Operation - struct for logs.
*/
type Operation struct {
	Code byte
	Body []byte
}

/*
Record - key-value struct in database.
*/
type Record struct {
	Key   string
	Value []byte
}
