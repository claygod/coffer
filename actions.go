package coffer

// Coffer
// Actions
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

// import (
// 	"github.com/claygod/coffer/usecases"
// )

func (c *Coffer) Write(key string, value []byte) error {
	return nil //TODO:
}

func (c *Coffer) WriteList(input map[string][]byte) error {
	return nil //TODO:
}

func (c *Coffer) Read(key string) ([]byte, error) {
	return nil, nil //TODO:
}

func (c *Coffer) ReadList(keys []string) (map[string][]byte, error) {
	return nil, nil //TODO:
}

func (c *Coffer) Delete(key string) error {
	return nil //TODO:
}

func (c *Coffer) DeleteList(keys []string) error {
	return nil //TODO:
}

func (c *Coffer) Transaction(handlerName string, keys []string, arg interface{}) error {
	return nil //TODO:
}
