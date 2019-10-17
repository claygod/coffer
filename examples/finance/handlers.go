package main

// Coffer
// Examples: finance handlers
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"unsafe"
)

// func HandlerNewAccount(arg interface{}, recs map[string][]byte) (map[string][]byte, error) {
// 	newAcc, ok := arg.(uint64)
// 	if !ok {
// 		return nil, fmt.Errorf("Invalid Argument: %v.", arg)
// 	} else if len(recs) != 1 {
// 		return nil, fmt.Errorf("Want 1 record, have %d", len(recs))
// 	}
// 	var recKey string
// 	for k, _ := range recs {
// 		recKey = k
// 	}
// 	return map[string][]byte{recKey: uint64ToBytes(newAcc)}, nil
// }

/*
HandlerCredit - credit handler.
*/
func HandlerCredit(arg []byte, recs map[string][]byte) (map[string][]byte, error) {
	if arg == nil || len(arg) != 8 {
		return nil, fmt.Errorf("Invalid Argument: %v.", arg)
	} else if len(recs) != 1 {
		return nil, fmt.Errorf("Want 1 record, have %d", len(recs))
	}
	delta := bytesToUint64(arg)
	var recKey string
	var recValue []byte
	for k, v := range recs {
		recKey = k
		recValue = v
	}
	if len(recValue) != 8 {
		return nil, fmt.Errorf("The length of the value in the record is %d bytes, but 8 bytes are needed", len(recValue))
	}
	curAcc := bytesToUint64(recValue)
	if delta > curAcc {
		return nil, fmt.Errorf("Not enough funds in the account. There is %d, a credit of %d.", curAcc, delta)
	}
	return map[string][]byte{recKey: uint64ToBytes(curAcc - delta)}, nil
}

/*
HandlerDebit - Debit handler.
*/
func HandlerDebit(arg []byte, recs map[string][]byte) (map[string][]byte, error) {
	if arg == nil || len(arg) != 8 {
		return nil, fmt.Errorf("Invalid Argument: %v.", arg)
	} else if len(recs) != 1 {
		return nil, fmt.Errorf("Want 1 record, have %d", len(recs))
	}
	delta := bytesToUint64(arg)
	var recKey string
	var recValue []byte
	for k, v := range recs {
		recKey = k
		recValue = v
	}
	if len(recValue) != 8 {
		return nil, fmt.Errorf("The length of the value in the record is %d bytes, but 8 bytes are needed", len(recValue))
	}
	curAmount := bytesToUint64(recValue)
	newAmount := curAmount + delta
	if curAmount > newAmount {
		return nil, fmt.Errorf("Account overflow. There is %d, a debit of %d.", curAmount, delta)
	}
	return map[string][]byte{recKey: uint64ToBytes(newAmount)}, nil
}

/*
HandlerTransfer - Transfer handler.
*/
func HandlerTransfer(arg []byte, recs map[string][]byte) (map[string][]byte, error) {
	if arg == nil {
		return nil, fmt.Errorf("Invalid Argument: %v.", arg)
	}
	dec := gob.NewDecoder(bytes.NewBuffer(arg))
	var req ReqTransfer
	if err := dec.Decode(&req); err != nil {
		return nil, fmt.Errorf("Invalid Argument: %v. Error: %v", arg, err)
	}

	if len(recs) != 2 {
		return nil, fmt.Errorf("Want 2 record, have %d", len(recs))
	}

	return helperHandlerTransfer(req, recs)
}

/*
HandlerMultiTransfer - MultiTransfer handler.
*/
func HandlerMultiTransfer(arg []byte, recs map[string][]byte) (map[string][]byte, error) {
	if arg == nil {
		return nil, fmt.Errorf("Invalid Argument: %v.", arg)
	}
	dec := gob.NewDecoder(bytes.NewBuffer(arg))
	var reqs []ReqTransfer
	if err := dec.Decode(&reqs); err != nil {
		return nil, fmt.Errorf("Invalid Argument: %v. Error: %v", string(arg), err)
	}

	if len(recs) != len(reqs)*2 {
		return nil, fmt.Errorf("Want %d record, have %d", len(reqs)*2, len(recs))
	}

	out := make(map[string][]byte, len(recs))
	for _, req := range reqs {
		vFrom, ok := recs[req.From]
		if !ok {
			return nil, fmt.Errorf("Entry %s cannot be found among transaction arguments.", req.From)
		}
		vTo, ok := recs[req.To]
		if !ok {
			return nil, fmt.Errorf("Entry %s cannot be found among transaction arguments.", req.To)
		}

		m, err := helperHandlerTransfer(req, map[string][]byte{req.From: vFrom, req.To: vTo})
		if err != nil {
			return nil, err
		}
		for k, v := range m {
			out[k] = v
		}
	}
	return out, nil
}

func helperHandlerTransfer(req ReqTransfer, recs map[string][]byte) (map[string][]byte, error) {
	recFromValueBytes, ok := recs[req.From]
	if !ok {
		return nil, fmt.Errorf("Entry %s cannot be found among transaction arguments.", req.From)
	} else if len(recFromValueBytes) != 8 {
		return nil, fmt.Errorf("The length of the value in the record `%s` is %d bytes, but 8 bytes are needed", req.From, len(recFromValueBytes))
	}
	recToValueBytes, ok := recs[req.To]
	if !ok {
		return nil, fmt.Errorf("Entry %s cannot be found among transaction arguments.", req.To)
	} else if len(recFromValueBytes) != 8 {
		return nil, fmt.Errorf("The length of the value in the record `%s` is %d bytes, but 8 bytes are needed", req.To, len(recToValueBytes))
	}

	recFromValueUint64 := bytesToUint64(recFromValueBytes)
	newAmountFrom := recFromValueUint64 - req.Amount
	if req.Amount > recFromValueUint64 {
		return nil, fmt.Errorf("Not enough funds in the account `%s`. There is %d, a credit of %d.", req.From, recFromValueUint64, req.Amount)
	}
	recToValueUint64 := bytesToUint64(recToValueBytes)
	newAmountTo := recToValueUint64 + req.Amount
	if recToValueUint64 > newAmountTo {
		return nil, fmt.Errorf("Account `%s` is overflow. There is %d, a debit of %d.", req.To, recToValueUint64, req.Amount)
	}

	return map[string][]byte{req.From: uint64ToBytes(newAmountFrom), req.To: uint64ToBytes(newAmountTo)}, nil
}

/*
ReqTransfer - request.
*/
type ReqTransfer struct {
	From   string
	To     string
	Amount uint64
}

func uint64ToBytes(i uint64) []byte {
	x := (*[8]byte)(unsafe.Pointer(&i))
	out := make([]byte, 0, 8)
	out = append(out, x[:]...)
	return out
}

func bytesToUint64(b []byte) uint64 {
	var x [8]byte
	copy(x[:], b[:])
	return *(*uint64)(unsafe.Pointer(&x))
}
