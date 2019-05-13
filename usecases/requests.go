package usecases

// Coffer
// Requests
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

type ReqWriteList struct {
	Time int64
	List map[string][]byte
}

type ReqLoadList struct {
	Time int64
	Keys []string
}

type ReqDeleteList struct {
	Time int64
	Keys []string
}

type ReqTransaction struct {
	Time        int64
	HandlerName string
	Keys        []string
	Value       interface{}
}
