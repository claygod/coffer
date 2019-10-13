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
	// defer fmt.Println("Время проведения оперерации ", time.Now().UnixNano()-tStart)

	rep := &reports.ReportTransaction{Report: reports.Report{}}
	// находим хандлер
	hdlx, err := t.handlers.Get(reqTr.HandlerName)
	if err != nil {
		rep.Code = codes.ErrHandlerNotFound
		rep.Error = err
		return rep
	}
	hdl := *hdlx
	// читаем текущие значения
	curRecsMap, notFound := repo.ReadList(reqTr.Keys)
	//curRecs, err := repo.GetRecords(reqTr.Keys)
	if len(notFound) != 0 {
		rep.Code = codes.ErrReadRecords
		rep.Error = fmt.Errorf("Records not found: %s", strings.Join(notFound, ", "))
		return rep
	}
	// проводим операцию  с полученными из репо значениями
	novRecsMap, err := hdl(reqTr.Value, curRecsMap)
	if err != nil {
		rep.Code = codes.ErrHandlerResponse
		rep.Error = err
		return rep
	}
	//сохранение изменённых записей (полученных в результате выполнения транзакции)
	repo.WriteList(novRecsMap)
	rep.Code = codes.Ok
	rep.Data = novRecsMap
	return rep
}
