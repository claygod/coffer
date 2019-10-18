package handlers

// Coffer
// Handlers repo
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"sync"

	"github.com/claygod/coffer/domain"
)

/*
Handlers - parallel storage
*/
type Handlers struct {
	mtx      sync.RWMutex
	handlers map[string]*domain.Handler
}

/*
New - create new Handlers.
*/
func New() *Handlers {
	return &Handlers{
		handlers: make(map[string]*domain.Handler),
	}
}

/*
Get - get record from storage.
*/
func (h *Handlers) Get(handlerName string) (*domain.Handler, error) {
	h.mtx.RLock()
	hdl, ok := h.handlers[handlerName]
	h.mtx.RUnlock()
	if !ok {
		return nil, fmt.Errorf("Header with the name `%s` is not installed.", handlerName)
	}
	return hdl, nil
}

/*
Set - add storage entry.
*/
func (h *Handlers) Set(handlerName string, handlerMethod *domain.Handler) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	// _, ok := h.handlers[handlerName]
	// if ok {
	// 	return fmt.Errorf("Header with the name `%s` is installed.", handlerName)
	// }
	h.handlers[handlerName] = handlerMethod
	//return nil
}
