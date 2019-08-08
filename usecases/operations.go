package usecases

// Coffer
// Operations helper
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/claygod/coffer/domain"
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
			if err := o.trn.doOperationTransaction(reqTr, repo); err != nil {
				return err
			}
		case codeDeleteList:
			reqDL, err := o.reqCoder.ReqDeleteListDecode(op.Body)
			if err != nil {
				return err
			} else if err := repo.DelList(reqDL.Keys); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unknown operation `%d`", op.Code)
		}
		//f.changesCounter += int64(len(op.Body)) //считаем в байтах
	}
	return nil
}

func (o *Operations) loadFromFile(filePath string) ([]*domain.Operation, error) {
	opFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer opFile.Close()
	fInfo, err := opFile.Stat()
	if err != nil {
		o.logger.Warning(err)
		return make([]*domain.Operation, 0), nil //тут можно и nil возвращать, но лучше всё же пустой список
	} else if fInfo.Size() == 0 {
		return make([]*domain.Operation, 0), nil //тут можно и nil возвращать, но лучше всё же пустой список
	}
	ops, err := o.loadOperationsFromFile(opFile)
	if err != nil {
		//TODO: тут логировать эту ошибку, т.к. она скорее warning
		o.logger.Warning(err)
	}
	return ops, nil
}

/*
loadOperationsFromFile - скачиваем операции из файла, возвращаемая ошибка
скорей всего означает, что какая-то операция не полностью была записана и невозможно
было её прочитать. Соответственно, ошибки не критические, и скорее нужны для логов.
*/
func (o *Operations) loadOperationsFromFile(fl *os.File) ([]*domain.Operation, error) {
	// st, _ := fl.Stat()
	// flSize := st.Size()
	counReadedBytes := 0
	ops := make([]*domain.Operation, 0, 16)
	rSize := make([]byte, 8)
	var errOut error
	for {
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
	if _, err := buf.Write(uint64ToBytes(uint64(len(op.Body) + 2))); err != nil { //TODO +1
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
