package usecases

// Coffer
// Transaction helper
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"strings"

	"github.com/claygod/coffer/domain"
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

func (t *Transaction) doOperationTransaction(reqTr *ReqTransaction, repo domain.RecordsRepository) error {
	// находим хандлер
	hdlx, err := t.handlers.Get(reqTr.HandlerName)
	if err != nil {
		return err
	}
	hdl := *hdlx
	// читаем текущие значения
	curRecsMap, notFound := repo.ReadList(reqTr.Keys)
	//curRecs, err := repo.GetRecords(reqTr.Keys)
	if len(notFound) != 0 {
		return fmt.Errorf("Records not found: %s", strings.Join(notFound, ", "))
	}
	// записи преобразуем в массив
	// curRecsMap := make(map[string][]byte)
	// for _, rec := range curRecs {
	// 	curRecsMap[rec.Key] = rec.Value
	// }
	// проводим операцию  с inmemory хранилищем
	novRecsMap, err := hdl(reqTr.Value, curRecsMap)
	if err != nil {
		return err
	}
	// массив преобразуем в список записей
	// novRecsList := make([]*domain.Record, 0, len(novRecsMap))
	// for key, value := range novRecsMap {
	// 	if _, ok := curRecsMap[key]; !ok {
	// 		// проверяем, чтобы хэндлер не натворил лишнего (не добавил новую запись)
	// 		return fmt.Errorf("Transaction Handler `%s` tries to change inaccessible records.", reqTr.HandlerName)
	// 	}
	// 	rec := &domain.Record{
	// 		Key:   key,
	// 		Value: value,
	// 	}
	// 	novRecsList = append(novRecsList, rec)
	// }
	// проверяем, нет ли надобности удалить какие-то записи
	delRecsList := make([]string, 0, len(reqTr.Keys))
	for _, key := range reqTr.Keys {
		if _, ok := novRecsMap[key]; !ok {
			delRecsList = append(delRecsList, key)
		}
	}
	//сохранение изменённых записей (полученных после выполнения транзакции)
	repo.WriteList(novRecsMap)
	// if err := repo.SetRecords(novRecsList); err != nil {
	// 	return err
	// }
	// удаление записей (при необходимости)
	if len(delRecsList) != 0 {
		repo.DelList(delRecsList)
	}
	return nil
}
