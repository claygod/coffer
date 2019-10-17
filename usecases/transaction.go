package usecases

// Coffer
// Transaction helper
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"strings"

	//"time"

	"github.com/claygod/coffer/domain"
	"github.com/claygod/coffer/reports"
	"github.com/claygod/coffer/reports/codes"
)

type Transaction struct {
	//repo domain.RecordsRepository
	handlers HandleStore
}

func NewTransaction(handlers HandleStore) *Transaction {
	return &Transaction{
		handlers: handlers,
	}
}

func (t *Transaction) doOperationTransaction(reqTr *ReqTransaction, repo domain.RecordsRepository) *reports.ReportTransaction {
	// tStart := time.Now().UnixNano()
	// defer fmt.Println("Operation time ", time.Now().UnixNano()-tStart)

	rep := &reports.ReportTransaction{Report: reports.Report{}}
	// find handler
	hdlx, err := t.handlers.Get(reqTr.HandlerName)
	if err != nil {
		rep.Code = codes.ErrHandlerNotFound
		rep.Error = err
		return rep
	}
	hdl := *hdlx
	// read the current values
	curRecsMap, notFound := repo.ReadList(reqTr.Keys)
	if len(notFound) != 0 {
		rep.Code = codes.ErrReadRecords
		rep.Error = fmt.Errorf("Records not found: %s", strings.Join(notFound, ", "))
		return rep
	}
	// we carry out the operation with the values obtained from the repo
	novRecsMap, err := hdl(reqTr.Value, curRecsMap)
	if err != nil {
		rep.Code = codes.ErrHandlerResponse
		rep.Error = err
		return rep
	}
	// saving modified records (obtained as a result of a transaction)
	repo.WriteList(novRecsMap)
	rep.Code = codes.Ok
	rep.Data = novRecsMap
	return rep
}
