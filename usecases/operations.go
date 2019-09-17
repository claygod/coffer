package usecases

// Coffer
// Operations helper
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/claygod/coffer/domain"
	"github.com/claygod/coffer/reports/codes"
)

type Operations struct {
	logger     Logger
	config     *Config
	reqCoder   *ReqCoder
	resControl Resourcer
	trn        *Transaction
}

func NewOperations(logger Logger, config *Config, reqCoder *ReqCoder, resControl Resourcer, trn *Transaction) *Operations {
	return &Operations{
		logger:     logger,
		config:     config,
		reqCoder:   reqCoder,
		resControl: resControl,
		trn:        trn,
	}
}

func (o *Operations) DoOperations(ops []*domain.Operation, repo domain.RecordsRepository) error {
	for _, op := range ops {
		if !o.resControl.GetPermission(int64(len(op.Body))) {
			return fmt.Errorf("Operation code %d, len(body)=%d, Not permission!", op.Code, len(op.Body))
		}
		//fmt.Println("Operation: ", string(op.Body))
		//TODO: пока не проверяем результаты операций, считаем, что раз он были ок в первый раз, должны быть ок и сейчас
		// если не ок, то надо всё останавливать, т.к. все записанные операции раньше были успешными
		switch op.Code {
		case codeWriteList:
			reqWL, err := o.reqCoder.ReqWriteListDecode(op.Body)
			if err != nil {
				return err
			}
			repo.WriteList(reqWL.List)
			// else if err := repo.SetRecords(o.convertReqWriteListToRecords(reqWL)); err != nil {
			// 	return err
			// }
		case codeTransaction:
			reqTr, err := o.reqCoder.ReqTransactionDecode(op.Body)
			if err != nil {
				return err
			}
			if rep := o.trn.doOperationTransaction(reqTr, repo); rep.Code != codes.Ok {
				return rep.Error
			}
		case codeDeleteListStrict:
			reqDL, err := o.reqCoder.ReqDeleteListDecode(op.Body)
			if err != nil {
				return err
			} else if notFound := repo.DelListStrict(reqDL.Keys); len(notFound) != 0 {
				//notFound := repo.DelListStrict(reqDL.Keys) //результат не важен, главное, что он такой же как и в предыдущий раз
				return fmt.Errorf("Operations:DoOperations:DeleteList:Keys not found: %s", strings.Join(notFound, ", "))
			}
		case codeDeleteListOptional:
			reqDL, err := o.reqCoder.ReqDeleteListDecode(op.Body)
			if err != nil {
				return err
			} else {
				repo.DelListOptional(reqDL.Keys)
			}
		default:
			return fmt.Errorf("Unknown operation `%d`", op.Code)
		}
		//f.changesCounter += int64(len(op.Body)) //считаем в байтах
	}
	return nil
}

func (o *Operations) loadFromFile(filePath string) ([]*domain.Operation, error, error) {
	opFile, err := os.Open(filePath)
	if err != nil {
		return nil, err, nil
	}
	defer opFile.Close()
	fInfo, err := opFile.Stat()
	if err != nil || fInfo.Size() == 0 {
		return make([]*domain.Operation, 0), nil, nil //тут можно и nil возвращать, но лучше всё же пустой список
	}
	ops, wrn := o.loadOperationsFromFile(opFile)
	return ops, nil, wrn
}

/*
loadOperationsFromFile - скачиваем операции из файла, возвращаемая ошибка
скорей всего означает, что какая-то операция не полностью была записана и невозможно
было её прочитать. Соответственно, ошибки не критические, и скорее нужны для логов.
(Так как критические, были бы при невозможности открыть файл, отсутствии директории,
а в данном случае в аргументах уже открытый файл, осталось его только прочитать.)
*/
func (o *Operations) loadOperationsFromFile(fl *os.File) ([]*domain.Operation, error) {
	// st, _ := fl.Stat()
	// flSize := st.Size()

	//stat, _ := fl.Stat()
	//fmt.Println("stst: ", stat.Size(), fl.Name())
	counReadedBytes := 0
	ops := make([]*domain.Operation, 0, 16)
	rSize := make([]byte, 8)
	var errOut error
	for {
		//fmt.Println("---loadOperationsFromFile")
		_, err := fl.Read(rSize)
		if err != nil {
			//fmt.Println("OP:LD:err1: ", err)
			if err != io.EOF {
				errOut = err //o.logger.Warning(err)
			}
			break
			//return nil, err
		}
		//fmt.Println("OP:LD:1:rSize: ", rSize)
		counReadedBytes += 8
		rSuint64 := bytesToUint64(rSize)
		//fmt.Println("OP:LD:1:rSuint64: ", rSuint64)
		bTotal := make([]byte, int(rSuint64))
		n, err := fl.Read(bTotal)
		if err != nil {
			// if err == io.EOF { // тут EOF не должно быть?????
			// break
			// }
			//fmt.Println("OP:LD:err2: ", err)
			errOut = err //o.logger.Warning(err)
			break
			//return nil, err
		} else if n != int(rSuint64) {
			errOut = fmt.Errorf("The operation is not fully loaded: %d from %d )", n, rSuint64)
			//o.logger.Warning(fmt.Errorf("The operation is not fully loaded: %d from %d )", n, rSuint64))
			//fmt.Println("OP:LD:err3: ", n, int(rSuint64), rSuint64)
			break
			//return nil, fmt.Errorf("The operation is not fully loaded: %d from %d )", n, rSuint64)
		}
		op, err := o.logToOperat(bTotal)
		if err != nil {
			errOut = err // o.logger.Warning(err)
			break
			//return nil, err
		}
		//fmt.Println("OP:LD:OP: ", op)
		ops = append(ops, op)
	}
	return ops, errOut
}

func (o *Operations) operatToLog(op *domain.Operation) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.Write(uint64ToBytes(uint64(len(op.Body) + 1))); err != nil { //TODO +1
		return nil, err
	}
	if err := buf.WriteByte(op.Code); err != nil {
		return nil, err
	}
	if _, err := buf.Write(op.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (o *Operations) logToOperat(in []byte) (*domain.Operation, error) {
	if len(in) < 3 { //TODO: разобраться с минимальной цифрой (через тесты)
		return nil, fmt.Errorf("Len of input operation array == %d", len(in))
	}
	op := &domain.Operation{
		Code: in[0],
		Body: in[1:],
	}
	return op, nil
}

// func (o *operations) convertReqWriteListToRecords(req *ReqWriteList) []*domain.Record {
// 	recs := make([]*domain.Record, 0, len(req.List))
// 	for key, value := range req.List {
// 		rec := &domain.Record{
// 			Key:   key,
// 			Value: value,
// 		}
// 		recs = append(recs, rec)
// 	}
// 	return recs
// }
