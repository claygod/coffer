package usecases

// Coffer
// Requests
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	"encoding/gob"
	"time"
)

type ReqCoder struct {
}

func NewReqCoder() *ReqCoder {
	return &ReqCoder{}
}

func (r *ReqCoder) ReqWriteListEncode(req *ReqWriteList) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(req)
	return buf.Bytes(), err
}

func (r *ReqCoder) ReqDeleteListEncode(req *ReqDeleteList) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(req)
	return buf.Bytes(), err
}

func (r *ReqCoder) ReqTransactionEncode(req *ReqTransaction) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(req)
	return buf.Bytes(), err
}

func (r *ReqCoder) ReqWriteListDecode(body []byte) (*ReqWriteList, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(body))
	var req ReqWriteList
	err := dec.Decode(&req)
	return &req, err
}

func (r *ReqCoder) ReqDeleteListDecode(body []byte) (*ReqDeleteList, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(body))
	var req ReqDeleteList
	err := dec.Decode(&req)
	return &req, err
}

func (r *ReqCoder) ReqTransactionDecode(body []byte) (*ReqTransaction, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(body))
	var req ReqTransaction
	err := dec.Decode(&req)
	return &req, err
}

type ReqWriteList struct {
	Time time.Time
	List map[string][]byte
}

type ReqLoadList struct {
	Time time.Time
	Keys []string
}

type ReqDeleteList struct {
	Time time.Time
	Keys []string
}

type ReqTransaction struct {
	Time        time.Time
	HandlerName string
	Keys        []string
	Value       interface{}
}
