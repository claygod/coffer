package usecases

// Coffer
// Requests
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	"encoding/gob"
	"time"
)

/*
ReqCoder - requests encoder and decoder.
*/
type ReqCoder struct {
}

/*
NewReqCoder - create new ReqCoder.
*/
func NewReqCoder() *ReqCoder {
	return &ReqCoder{}
}

/*
ReqWriteListEncode - encode ReqWriteList
*/
func (r *ReqCoder) ReqWriteListEncode(req *ReqWriteList) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(req)

	return buf.Bytes(), err
}

/*
ReqDeleteListEncode - encode ReqDeleteList
*/
func (r *ReqCoder) ReqDeleteListEncode(req *ReqDeleteList) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(req)

	return buf.Bytes(), err
}

/*
ReqTransactionEncode - encode ReqTransaction
*/
func (r *ReqCoder) ReqTransactionEncode(req *ReqTransaction) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(req)

	return buf.Bytes(), err
}

/*
ReqWriteListDecode - decode ReqWriteList
*/
func (r *ReqCoder) ReqWriteListDecode(body []byte) (*ReqWriteList, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(body))
	var req ReqWriteList
	err := dec.Decode(&req)

	return &req, err
}

/*
ReqDeleteListDecode - decode ReqDeleteList
*/
func (r *ReqCoder) ReqDeleteListDecode(body []byte) (*ReqDeleteList, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(body))
	var req ReqDeleteList
	err := dec.Decode(&req)

	return &req, err
}

/*
ReqTransactionDecode - decode ReqTransaction
*/
func (r *ReqCoder) ReqTransactionDecode(body []byte) (*ReqTransaction, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(body))
	var req ReqTransaction
	err := dec.Decode(&req)

	return &req, err
}

/*
ReqWriteList - write list request
*/
type ReqWriteList struct {
	Time time.Time
	List map[string][]byte
}

/*
ReqLoadList - load list  request
*/
type ReqLoadList struct {
	Time time.Time
	Keys []string
}

/*
ReqDeleteList - delete list request
*/
type ReqDeleteList struct {
	Time time.Time
	Keys []string
}

/*
ReqTransaction - transaction request
*/
type ReqTransaction struct {
	Time        time.Time
	HandlerName string
	Keys        []string
	Value       []byte
}
