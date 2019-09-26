package coffer

// Coffer
// API tests
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/claygod/coffer/domain"
	// "github.com/claygod/coffer/services"
	// "github.com/claygod/coffer/services/filenamer"
	// "github.com/claygod/coffer/services/journal"
	// "github.com/claygod/coffer/services/repositories/records"
	// "github.com/claygod/coffer/reports"
	"github.com/claygod/coffer/reports/codes"
	"github.com/claygod/coffer/services/journal"
	"github.com/claygod/coffer/services/repositories/handlers"
	"github.com/claygod/coffer/services/resources"

	// "github.com/claygod/coffer/services/startstop"
	"github.com/claygod/coffer/usecases"
	// "github.com/claygod/tools/logger"
	// "github.com/claygod/tools/porter"
)

const dirPath string = "./test/"

func TestCofferCleanDir(t *testing.T) {
	forTestClearDir(dirPath)
}

func TestCofferTransaction(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	// hdlExch := domain.Handler(handlerExchange)
	// cof1.SetHandler("exchange", &hdlExch)
	if !cof1.Start() {
		t.Error("Could not start the application (1)")
		return
	} else {
		defer cof1.Stop()
	}
	cof1.Write("aaa", []byte("111"))
	cof1.Write("bbb", []byte("222"))
	if rep := cof1.Transaction("exchange", []string{"aaa", "bbb"}, nil); rep.Code >= codes.Warning {
		t.Error(err)
		return
	}
	// количество записей не должно измениться
	if rep := cof1.Count(); rep.Count != 2 {
		t.Errorf("Records (cof1) count, have %d, want 2.", rep.Count)
		return
	}
	// количество записей не должно измениться
	rep := cof1.ReadList([]string{"aaa", "bbb"})
	if rep.Code >= codes.Warning {
		t.Errorf("Transaction results: code=%d , data=%v, not_found=%v, err=%v.", rep.Code, rep.Data, rep.NotFound, rep.Error)
		return
	} else if string(rep.Data["aaa"]) != "222" || string(rep.Data["bbb"]) != "111" {
		t.Errorf("Want aaa==222 bbb==111 , have aaa=%s bbb==%s ", string(rep.Data["aaa"]), string(rep.Data["bbb"]))
	}
}

func TestCofferTransactionRecordsNotFound(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	// hdlExch := domain.Handler(handlerExchange)
	// cof1.SetHandler("exchange", &hdlExch)
	if !cof1.Start() {
		t.Error("Could not start the application (1)")
		return
	} else {
		defer cof1.Stop()
	}
	cof1.Write("aaa", []byte("111"))
	cof1.Write("bbb", []byte("222"))
	if rep := cof1.Transaction("exchange", []string{"xxx", "yyy"}, nil); rep.Code != codes.ErrReadRecords {
		t.Errorf("Want codes.ErrReadRecords , have %v", rep.Code)
		t.Error(rep.Error)
		return
	}
}

func TestCofferTransactionRecordsBigLenKeys(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err, wrn := createNewCofferLength4T(2, 10)
	if err != nil {
		t.Error(err)
		return
	} else if wrn != nil {
		t.Error(wrn)
		return
	}
	// hdlExch := domain.Handler(handlerExchange)
	// cof1.SetHandler("exchange", &hdlExch)
	if !cof1.Start() {
		t.Error("Could not start the application (1)")
		return
	} else {
		defer cof1.Stop()
	}
	if rep := cof1.Transaction("exchange", []string{"xxxxx"}, nil); rep.Code != codes.ErrExceedingMaxKeyLength {
		t.Errorf("Want codes.ErrExceedingMaxKeyLength , have %v", rep.Code)
		return
	}
}

func TestCofferTransactionRecordsBigOperationsCount(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err, wrn := createNewCofferLength4T(2, 10)
	if err != nil {
		t.Error(err)
		return
	} else if wrn != nil {
		t.Error(wrn)
		return
	}
	// hdlExch := domain.Handler(handlerExchange)
	// cof1.SetHandler("exchange", &hdlExch)
	if !cof1.Start() {
		t.Error("Could not start the application (1)")
		return
	} else {
		defer cof1.Stop()
	}
	if rep := cof1.Transaction("exchange", []string{"x", "x", "x", "x", "x", "x", "x",
		"x", "x", "x", "x", "x", "x", "x"}, nil); rep.Code != codes.ErrRecordLimitExceeded {
		t.Errorf("Want codes.ErrRecordLimitExceeded , have %v", rep.Code)
		return
	}
}

func TestCofferTransactionNotFound(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err, wrn := createNewCofferT()
	if err != nil {
		t.Error(err)
		return
	} else if wrn != nil {
		t.Error(wrn)
		return
	}
	if !cof1.Start() {
		t.Error("Could not start the application (1)")
		return
	} else {
		defer cof1.Stop()
	}
	if rep := cof1.Transaction("exchangeXXX", []string{"aaa", "bbb"}, nil); rep.Code != codes.ErrHandlerNotFound {
		t.Error("Handler is not available, but for some reason is executed.")
		return
	}
}

func TestCofferStartStop(t *testing.T) {
	forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferLength(t, 4, 7)
	if err != nil {
		t.Error(err)
		return
	}
	if err := cof1.Save(); err != nil {
		t.Error(err)
	}
	if !cof1.Start() {
		t.Error("Could not start the application (1)")
	}
	if !cof1.Start() {
		t.Error("Could not start the application (2)")
	}
	if !cof1.Stop() {
		t.Error("Could not stop the application (1)")
	}
	if !cof1.Stop() {
		t.Error("Could not stop the application (2)")
	}
	if err := cof1.Save(); err != nil {
		t.Error(err)
	}
}

func TestCofferStopHard(t *testing.T) {
	forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferLength(t, 4, 7)
	if err != nil {
		t.Error(err)
		return
	}
	if !cof1.Start() {
		t.Error("Could not start the application (1)")
	}
	if err := cof1.StopHard(); err != nil {
		t.Error(err)
	}
}

func TestCofferWriteRead(t *testing.T) {
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferLength(t, 4, 7)
	if err != nil {
		t.Error(err)
		return
	}
	// без ошибок
	t.Log("Stage1")
	if rep := cof1.Write("aa", []byte("bbb")); rep.Code != codes.Ok || rep.Error != nil {
		t.Error("Write: want code 0 (Ok), have code: ", rep.Code, " Resp. err.: ", rep.Error)
		return
	}
	if rep := cof1.Read("aa"); rep.Code != codes.Ok || rep.Error != nil || rep.Data == nil {
		t.Error("Read: want code 0 (Ok), have code: ", rep.Code, " Resp. err.: ", rep.Error, " Resp. data: ", rep.Data)
		return
	}
	// -- пишем слишком большой ключ
	t.Log("Stage2")
	if rep := cof1.Write("aaaaa", []byte("bbb")); rep.Code != codes.ErrExceedingMaxKeyLength || rep.Error == nil {
		t.Error("Write: want code `ErrExceedingMaxKeyLength`, have code: ", rep.Code, " Resp. err.: ", rep.Error)
		return
	}
	// -- пишем слишком большое значение
	t.Log("Stage3")
	if rep := cof1.Write("dd", []byte("cccccccccccc")); rep.Code != codes.ErrExceedingMaxValueSize || rep.Error == nil {
		t.Error("Write: want code `ErrExceedingMaxValueSize`, have code: ", rep.Code, " Resp. err.: ", rep.Error)
		return
	}
	// -- пытаемся считать несуществующую запись
	t.Log("Stage4")
	if rep := cof1.Read("xx"); rep.Code != codes.ErrReadRecords || rep.Error != nil || rep.Data != nil {
		t.Error("Read: want code `ErrReadRecords`, have code: ", rep.Code, " Resp. err.: ", rep.Error, " Resp. data: ", rep.Data)
		return
	}
}

func TestCofferWriteListReadList(t *testing.T) {
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}
	req := make(map[string][]byte)
	for i := 10; i < 19; i++ {
		req["aasa"+strconv.Itoa(i)] = []byte("bbsb" + strconv.Itoa(i))
	}
	// -- записываем список
	t.Log("Stage1")
	if rep := cof1.WriteList(req); rep.Code >= codes.Warning || rep.Error != nil {
		t.Error(err)
	}
	if rep := cof1.Count(); rep.Count != 9 {
		t.Errorf("Records (cof1) count, have %d, wand 9.", rep.Count)
		return
	}
	// -- считываем реальный список
	t.Log("Stage2")
	rep := cof1.ReadList([]string{"aasa10", "aasa11"})
	if rep.Code != codes.Ok || rep.Error != nil || len(rep.Data) != 2 || len(rep.NotFound) != 0 {
		t.Error(err)
		t.Error(rep)
		return
	}
	v, ok := rep.Data["aasa10"]
	if !ok {
		t.Error("Key `aasa10` not found")
		return
	} else if string(v) != "bbsb10" {
		t.Errorf("Key `aasa10`: value want `bbsb10` have %s", string(v))
		return
	}
	// -- пытаемся считать и несуществующие записи тоже
	t.Log("Stage3")
	rep = cof1.ReadList([]string{"aasa10", "aasa90"})
	if rep.Code != codes.ErrReadRecords || rep.Error != nil || len(rep.Data) != 1 || len(rep.NotFound) != 1 {
		t.Error(err)
		t.Error(rep)
		return
	}
	v, ok = rep.Data["aasa10"]
	if !ok {
		t.Error("Key `aasa10` not found")
		return
	} else if string(v) != "bbsb10" {
		t.Errorf("Key `aasa10`: value want `bbsb10` have %s", string(v))
		return
	} else if rep.NotFound[0] != "aasa90" {
		t.Errorf("Not found: want `aasa90` have %s", rep.NotFound[0])
	}
	cof1.Stop()
}

func TestCofferKeyLength(t *testing.T) {
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferLength(t, 3, 7)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Stage1")
	if rep := cof1.Write("aa1aa", []byte("bbsb10")); rep.Code != codes.ErrExceedingMaxKeyLength || rep.Error == nil {
		t.Error(codes.ErrExceedingMaxKeyLength, rep, err)
		return
	} else if rep := cof1.Write("aa1", []byte("bb1bbbbbbbbbbbbb")); rep.Code != codes.ErrExceedingMaxValueSize || rep.Error == nil {
		t.Error(codes.ErrExceedingMaxValueSize, rep, err)
		return
	}
	t.Log("Stage2")
	if rep := cof1.Write("aa1", []byte("bb1")); rep.Code != codes.Ok || rep.Error != nil {
		t.Error(codes.Ok, rep, err)
		return
	} else if rep := cof1.Read("aa1"); rep.Code != codes.Ok || rep.Error != nil {
		t.Error(rep, err)
		return
	} else if rep := cof1.Read("aa1aa"); rep.Code != codes.ErrExceedingMaxKeyLength || rep.Error == nil {
		t.Error(codes.ErrExceedingMaxKeyLength, rep, err)
		return
	}
	t.Log("Stage3")
	if rep := cof1.Delete("aa1aa"); rep.Code != codes.ErrExceedingMaxKeyLength || rep.Error == nil {
		t.Error(codes.ErrExceedingMaxKeyLength, rep, err)
		return

	}
}

func TestCofferMaxCountPerOperation(t *testing.T) {
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}
	// попытка записать за один раз слишком много записей
	t.Log("Stage1")
	reqWriteList := make(map[string][]byte)
	reqReadList := make([]string, 0)
	for i := 10; i < 22; i++ {
		reqWriteList["aasa"+strconv.Itoa(i)] = []byte("bbsb" + strconv.Itoa(i))
		reqReadList = append(reqReadList, "aasa"+strconv.Itoa(i))
	}
	rep := cof1.WriteList(reqWriteList)
	if rep.Code != codes.ErrRecordLimitExceeded || rep.Error == nil {
		t.Errorf("Want `ErrRecordLimitExceeded` have code `%d`", rep.Code)
		t.Error(rep.Error)
		return
	}
	t.Log(rep)
	// попытка прочитать за один раз слишком много записей
	t.Log("Stage2")
	rep2 := cof1.ReadList(reqReadList)
	if rep2.Code != codes.ErrRecordLimitExceeded || rep2.Error == nil {
		t.Errorf("Want `ErrRecordLimitExceeded` have code `%d`", rep2.Code)
		t.Error(rep2.Error)
		return
	}
	// попытка удалить Strict за один раз слишком много записей
	t.Log("Stage3")
	rep3 := cof1.DeleteListStrict(reqReadList)
	if rep3.Code != codes.ErrRecordLimitExceeded || rep3.Error == nil {
		t.Errorf("Want `ErrRecordLimitExceeded` have code `%d`", rep2.Code)
		t.Error(rep3.Error)
		return
	}
	// попытка удалить Optional за один раз слишком много записей
	t.Log("Stage3")
	rep3 = cof1.DeleteListOptional(reqReadList)
	if rep3.Code != codes.ErrRecordLimitExceeded || rep3.Error == nil {
		t.Errorf("Want `ErrRecordLimitExceeded` have code `%d`", rep2.Code)
		t.Error(rep3.Error)
		return
	}
}

func TestCofferLoadFromLogs(t *testing.T) {
	defer forTestClearDir(dirPath)

	// наполняем базу и сохраняем в память её логи
	t.Log("Stage1")
	cof1, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}
	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb")); rep.Code > codes.Warning || rep.Error != nil {
			t.Error(err)
		}
		//time.Sleep(100 * time.Millisecond)
	}
	if rep := cof1.Count(); rep.Count != 9 {
		t.Errorf("Records (cof1) count, have %d, wand 9.", rep.Count)
		return
	}
	cof1.Stop()
	b1, err := ioutil.ReadFile(dirPath + "1.log") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	b2, err := ioutil.ReadFile(dirPath + "2.log") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	//fmt.Println("=> ", string(b1))
	//fmt.Println("=> ", string(b2))

	// пробуем загрузиться с логов
	t.Log("Stage2")
	cof111, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}
	if rep := cof111.Count(); rep.Count != 9 {
		t.Errorf("Records (cof111) count, have %d, wand 9.", rep.Count)
		return
	} else {
		t.Log("Load true logs OK")
	}
	cof111.Stop()
	forTestClearDir(dirPath)

	// специально портим один файл, и одна запись в нём при скачке должна быть потеряна
	t.Log("Stage3")
	if err := ioutil.WriteFile(dirPath+"3.log", b1, os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	if err := ioutil.WriteFile(dirPath+"4.log", b2[:len(b2)-1], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}
	if rep := cof2.Count(); rep.Count != 8 { // одна запись поломана и её нет, а почему-то скачена
		t.Errorf("Records (cof2) count, have %d, want 8.", rep.Count)
		return
	} else {
		t.Log("Load false logs OK")
	}
	cof2.Stop()
	forTestClearDir(dirPath)

	// переименовываем один файл, в результате получив нормальный после битого
	// но этот последний файл не должен быть загружен, т.к. загрузка должна остановиться на нём
	t.Log("Stage4")
	if err := ioutil.WriteFile(dirPath+"9.log", b1, os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	if err := ioutil.WriteFile(dirPath+"8.log", b2[:len(b2)-1], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	_, err, wrn := createNewCoffer()
	if err == nil {
		t.Error("Want error (The spoiled log...)")
		return
	} else {
		t.Log(wrn)
		t.Log("Load false/true logs OK")
	}
	//time.Sleep(15000 * time.Millisecond)
}

func TestCofferLoadFromLogsTransaction(t *testing.T) {
	defer forTestClearDir(dirPath)

	// наполняем базу и сохраняем в память её логи
	t.Log("Stage1")
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	// hdlExch := domain.Handler(handlerExchange)
	// cof1.SetHandler("exchange", &hdlExch)
	//cof1.Start()

	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb"+strconv.Itoa(i))); rep.Code > codes.Warning || rep.Error != nil {
			t.Error(err)
		}
		//time.Sleep(100 * time.Millisecond)
	}
	if rep := cof1.Transaction("exchange", []string{"aasa10", "aasa11"}, nil); rep.Code >= codes.Warning {
		t.Error(rep)
		return
	}

	if rep := cof1.Count(); rep.Count != 9 {
		t.Errorf("Records (cof1) count, have %d, wand 9.", rep.Count)
		return
	}
	cof1.Stop()
	b1, err := ioutil.ReadFile(dirPath + "1.log") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	b2, err := ioutil.ReadFile(dirPath + "2.log") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	//fmt.Println("=> ", string(b1))
	//fmt.Println("=> ", string(b2))

	// пробуем загрузиться с логов
	t.Log("Stage2")
	cof111, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	if rep := cof111.Count(); rep.Count != 9 {
		t.Errorf("Records (cof111) count, have %d, wand 9.", rep.Count)
		return
	} else {
		t.Log("Load true logs OK")
	}
	if rep := cof111.Read("aasa10"); string(rep.Data) != "bbsb11" {
		t.Errorf("Record have %s, wand `bbsb11`.", string(rep.Data))
		return
	}
	cof111.Stop()
	forTestClearDir(dirPath)

	// специально портим один файл, и одна запись в нём при скачке должна быть потеряна
	t.Log("Stage3")
	if err := ioutil.WriteFile(dirPath+"3.log", b1, os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	if err := ioutil.WriteFile(dirPath+"4.log", b2[:len(b2)-1], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	if rep := cof2.Count(); rep.Count != 9 { // одна запись поломана и её нет, а почему-то скачена
		t.Errorf("Records (cof2) count, have %d, want 9.", rep.Count)
		return
	} else {
		t.Log("Load false logs OK")
	}
	if rep := cof2.Read("aasa10"); string(rep.Data) != "bbsb10" { // теперь последнее действие (транзакция) отменена
		t.Errorf("Record have %s, wand `bbsb10`.", string(rep.Data))
		return
	}
	cof2.Stop()
	forTestClearDir(dirPath)

	// переименовываем один файл, в результате получив нормальный после битого
	// но этот последний файл не должен быть загружен, т.к. загрузка должна остановиться на нём
	t.Log("Stage4")
	if err := ioutil.WriteFile(dirPath+"9.log", b1, os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	if err := ioutil.WriteFile(dirPath+"8.log", b2[:len(b2)-1], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	_, err, wrn := createNewCofferT()
	if err == nil {
		t.Error("Want error (The spoiled log...)")
		return
	} else {
		t.Log(wrn)
		t.Log("Load false/true logs OK")
	}
	//time.Sleep(15000 * time.Millisecond)
}

func TestCofferLoadFromCheckpoint(t *testing.T) {
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}
	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb")); rep.Code > codes.Warning || rep.Error != nil {
			t.Error(err)
		}
		//time.Sleep(10 * time.Millisecond)
	}
	cof1.Stop()
	//time.Sleep(5 * time.Millisecond)
	b1, err := ioutil.ReadFile(dirPath + "3.checkpoint") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	forTestClearDir(dirPath)
	// проверяем загрузку с нормального, небитого файла
	t.Log("Stage1")
	if err := ioutil.WriteFile(dirPath+"3.checkpoint", b1, os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}

	if rep := cof2.Count(); rep.Count != 9 { // не все записи скачены
		t.Errorf("Records (cof2) count, have %d, wand 9.", rep.Count)
		return
	}
	cof2.Stop()
	forTestClearDir(dirPath)
	// проверяем загрузку с битого файла
	t.Log("Stage2")
	if err := ioutil.WriteFile(dirPath+"3.checkpoint", b1[:len(b1)-2], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof3, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}

	if rep := cof3.Count(); rep.Count != 0 { // все записи битого чекпоинта должны быть проигнорированы
		t.Errorf("Records (cof3) count, have %d, wand 0.", rep.Count)
		return
	}
}

// func TestCofferLoadFromCheckpointTransaction(t *testing.T) {
// 	defer forTestClearDir(dirPath)
// 	cof1, err, wrn := createNewCoffer()
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	} else if wrn != nil {
// 		t.Error(wrn)
// 		return
// 	}
// 	hdlExch := domain.Handler(handlerExchange)
// 	cof1.SetHandler("exchange", &hdlExch)
// 	cof1.Start()
// 	for i := 10; i < 19; i++ {
// 		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb")); rep.Code > codes.Warning || rep.Error != nil {
// 			t.Error(err)
// 		}
// 		//time.Sleep(10 * time.Millisecond)
// 	}
// 	if rep := cof1.Transaction("exchange", []string{"aasa10", "aasa11"}, nil); rep.Code >= codes.Warning {
// 		t.Error(rep)
// 		return
// 	}
// 	cof1.Stop()
// 	time.Sleep(5 * time.Millisecond)
// 	b1, err := ioutil.ReadFile(dirPath + "3.checkpoint") // сохраняем в память
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	forTestClearDir(dirPath)
// 	// проверяем загрузку с нормального, небитого файла
// 	t.Log("Stage1")
// 	if err := ioutil.WriteFile(dirPath+"3.checkpoint", b1, os.ModePerm); err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	cof2, err := createAndStartNewCoffer(t)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	if rep := cof2.Count(); rep.Count != 9 { // не все записи скачены
// 		t.Errorf("Records (cof2) count, have %d, wand 9.", rep.Count)
// 		return
// 	}
// 	cof2.Stop()
// 	forTestClearDir(dirPath)
// 	// проверяем загрузку с битого файла
// 	t.Log("Stage2")
// 	if err := ioutil.WriteFile(dirPath+"3.checkpoint", b1[:len(b1)-2], os.ModePerm); err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	cof3, err := createAndStartNewCoffer(t)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	if rep := cof3.Count(); rep.Count != 0 { // все записи битого чекпоинта должны быть проигнорированы
// 		t.Errorf("Records (cof3) count, have %d, wand 0.", rep.Count)
// 		return
// 	}
// }

func TestCofferLoadFromFalseCheckpointTrueLogs(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err, wrn := createNewCoffer()
	if err != nil {
		t.Error(err)
		return
	} else if wrn != nil {
		t.Error(wrn)
		return
	}
	//hdlExch := domain.Handler(handlerExchange)
	//cof1.SetHandler("exchange", &hdlExch)
	cof1.Start()
	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb"+strconv.Itoa(i))); rep.Code > codes.Warning || rep.Error != nil {
			t.Error(err)
		}
		//time.Sleep(10 * time.Millisecond)
	}
	// if rep := cof1.Transaction("exchange", []string{"aasa10", "aasa11"}, nil); rep.Code >= codes.Warning {
	// 	t.Error(rep)
	// 	return
	// }
	cof1.Stop()
	//time.Sleep(1 * time.Second)
	b1, err := ioutil.ReadFile(dirPath + "3.checkpoint") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	// проверяем загрузку с нормального, небитого файла
	t.Log("Stage1")
	if err := ioutil.WriteFile(dirPath+"3.checkpoint", b1[:len(b1)-2], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}

	if rep := cof2.Count(); rep.Count != 9 { // не все записи скачены, хотя при битом чекпоинте всё должно было быть скачено с логов
		t.Errorf("Records (cof2) count, have %d, wand 9.", rep.Count)
		return
	}
	cof2.Stop()
}

func TestCofferLoadFromFalseCheckpointTrueLogsTransaction(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	// hdlExch := domain.Handler(handlerExchange)
	// if err := cof1.SetHandler("exchange", &hdlExch); err != nil {
	// 	t.Error(wrn)
	// 	return
	// }
	//cof1.Start()
	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb"+strconv.Itoa(i))); rep.Code > codes.Warning || rep.Error != nil {
			t.Error(err)
		}
		//time.Sleep(10 * time.Millisecond)
	}
	if rep := cof1.Transaction("exchange", []string{"aasa10", "aasa11"}, nil); rep.Code >= codes.Warning {
		t.Error(rep)
		return
	}
	cof1.Stop()
	//time.Sleep(5 * time.Second)
	b1, err := ioutil.ReadFile(dirPath + "3.checkpoint") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	// проверяем загрузку с нормального, небитого файла
	t.Log("Stage1")
	if err := ioutil.WriteFile(dirPath+"3.checkpoint", b1[:len(b1)-2], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	//hdlExch := domain.Handler(handlerExchange)
	// if err := cof2.SetHandler("exchange", &hdlExch); err != nil {
	// 	t.Error(wrn)
	// 	return
	// }

	if rep := cof2.Count(); rep.Count != 9 { // не все записи скачены, хотя при битом чекпоинте всё должно было быть скачено с логов
		t.Errorf("Records (cof2) count, have %d, wand 9.", rep.Count)
		return
	}
	cof2.Stop()
}

// =======================================================================
// =========================== HELPERS ===================================
// =======================================================================

func createAndStartNewCoffer(t *testing.T) (*Coffer, error) {
	cof1, err, wrn := createNewCoffer()
	if err != nil {
		//fmt.Println("++1++", err)
		return nil, err
	} else if wrn != nil {
		t.Log(wrn)
	}
	if !cof1.Start() {
		//fmt.Println("++2++")
		return nil, fmt.Errorf("Failed to start (cof)")
	}
	return cof1, nil
}

func createAndStartNewCofferT(t *testing.T) (*Coffer, error) {
	cof1, err, wrn := createNewCofferT()
	if err != nil {
		//fmt.Println("++1++", err)
		return nil, err
	} else if wrn != nil {
		t.Log(wrn)
	}
	if !cof1.Start() {
		//fmt.Println("++2++")
		return nil, fmt.Errorf("Failed to start (cof)")
	}
	return cof1, nil
}

func createAndStartNewCofferLength(t *testing.T, maxKeyLength int, maxValueLength int) (*Coffer, error) {
	cof1, err, wrn := createNewCofferLength4(maxKeyLength, maxValueLength)
	if err != nil {
		return nil, err
	} else if wrn != nil {
		t.Log(wrn)
	}
	if !cof1.Start() {
		return nil, fmt.Errorf("Failed to start (cof)")
	}
	return cof1, nil
}

func createAndStartNewCofferLengthB(t *testing.B, maxKeyLength int, maxValueLength int) (*Coffer, error) {
	cof1, err, wrn := createNewCofferLength4(maxKeyLength, maxValueLength)
	if err != nil {
		return nil, err
	} else if wrn != nil {
		t.Log(wrn)
	}
	if !cof1.Start() {
		return nil, fmt.Errorf("Failed to start (cof)")
	}
	return cof1, nil
}

func createNewCoffer() (*Coffer, error, error) {
	jCnf := &journal.Config{
		BatchSize:              2000,
		LimitRecordsPerLogfile: 5,
	}
	ucCnf := &usecases.Config{
		FollowPause:             400 * time.Millisecond,
		LogsByCheckpoint:        2,
		DirPath:                 dirPath, // "/home/ed/goPath/src/github.com/claygod/coffer/test",
		AllowStartupErrLoadLogs: true,
		MaxKeyLength:            100,
		MaxValueLength:          10000,
		RemoveUnlessLogs:        true, // чистим логи после использования
	}
	rcCnf := &resources.Config{
		LimitMemory: 1000 * megabyte, // minimum available memory (bytes)
		LimitDisk:   1000 * megabyte, // minimum free disk space
		DirPath:     dirPath,         // "/home/ed/goPath/src/github.com/claygod/coffer/test"
	}

	cnf := &Config{
		JournalConfig:       jCnf,
		UsecasesConfig:      ucCnf,
		ResourcesConfig:     rcCnf,
		MaxRecsPerOperation: 10,
		//MaxKeyLength:        100,
		//MaxValueLength:      10000,
	}
	return New(cnf, nil)
}

func createNewCofferT() (*Coffer, error, error) {
	jCnf := &journal.Config{
		BatchSize:              2000,
		LimitRecordsPerLogfile: 5,
	}
	ucCnf := &usecases.Config{
		FollowPause:             400 * time.Millisecond,
		LogsByCheckpoint:        2,
		DirPath:                 dirPath, // "/home/ed/goPath/src/github.com/claygod/coffer/test",
		AllowStartupErrLoadLogs: true,
		MaxKeyLength:            100,
		MaxValueLength:          10000,
		RemoveUnlessLogs:        true, // чистим логи после использования
	}
	rcCnf := &resources.Config{
		LimitMemory: 1000 * megabyte, // minimum available memory (bytes)
		LimitDisk:   1000 * megabyte, // minimum free disk space
		DirPath:     dirPath,         // "/home/ed/goPath/src/github.com/claygod/coffer/test"
	}

	cnf := &Config{
		JournalConfig:       jCnf,
		UsecasesConfig:      ucCnf,
		ResourcesConfig:     rcCnf,
		MaxRecsPerOperation: 10,
		//MaxKeyLength:        100,
		//MaxValueLength:      10000,
	}

	hdlExch := domain.Handler(handlerExchange)
	hdls := handlers.New()
	hdls.Set("exchange", &hdlExch)
	return New(cnf, hdls)
}

func createNewCofferLength4(maxKeyLength int, maxValueLength int) (*Coffer, error, error) {
	jCnf := &journal.Config{
		BatchSize:              2000,
		LimitRecordsPerLogfile: 5,
	}
	ucCnf := &usecases.Config{
		FollowPause:             400 * time.Millisecond,
		LogsByCheckpoint:        2,
		DirPath:                 dirPath, // "/home/ed/goPath/src/github.com/claygod/coffer/test",
		AllowStartupErrLoadLogs: true,
		MaxKeyLength:            maxKeyLength,
		MaxValueLength:          maxValueLength,
		RemoveUnlessLogs:        true, // чистим логи после использования
	}
	rcCnf := &resources.Config{
		LimitMemory: 1000 * megabyte, // minimum available memory (bytes)
		LimitDisk:   1000 * megabyte, // minimum free disk space
		DirPath:     dirPath,         // "/home/ed/goPath/src/github.com/claygod/coffer/test"
	}

	cnf := &Config{
		JournalConfig:       jCnf,
		UsecasesConfig:      ucCnf,
		ResourcesConfig:     rcCnf,
		MaxRecsPerOperation: 10,
		//MaxKeyLength:        100,
		//MaxValueLength:      10000,
	}
	return New(cnf, nil)
}

func createNewCofferLength4T(maxKeyLength int, maxValueLength int) (*Coffer, error, error) {
	jCnf := &journal.Config{
		BatchSize:              2000,
		LimitRecordsPerLogfile: 5,
	}
	ucCnf := &usecases.Config{
		FollowPause:             400 * time.Millisecond,
		LogsByCheckpoint:        2,
		DirPath:                 dirPath, // "/home/ed/goPath/src/github.com/claygod/coffer/test",
		AllowStartupErrLoadLogs: true,
		MaxKeyLength:            maxKeyLength,
		MaxValueLength:          maxValueLength,
		RemoveUnlessLogs:        true, // чистим логи после использования
	}
	rcCnf := &resources.Config{
		LimitMemory: 1000 * megabyte, // minimum available memory (bytes)
		LimitDisk:   1000 * megabyte, // minimum free disk space
		DirPath:     dirPath,         // "/home/ed/goPath/src/github.com/claygod/coffer/test"
	}

	cnf := &Config{
		JournalConfig:       jCnf,
		UsecasesConfig:      ucCnf,
		ResourcesConfig:     rcCnf,
		MaxRecsPerOperation: 10,
		//MaxKeyLength:        100,
		//MaxValueLength:      10000,
	}
	hdlExch := domain.Handler(handlerExchange)
	hdls := handlers.New()
	hdls.Set("exchange", &hdlExch)
	return New(cnf, hdls)
}

func forTestClearDir(dir string) error {
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		//fmt.Println(name)
		if strings.HasSuffix(name, ".log") || strings.HasSuffix(name, ".check") || strings.HasSuffix(name, ".checkpoint") {
			os.Remove(dir + name)
		}
		//		err = os.RemoveAll(filepath.Join(dir, name))
		//		if err != nil {
		//			return err
		//		}
	}
	return nil
}

func handlerExchange(arg interface{}, recs map[string][]byte) (map[string][]byte, error) {
	if arg != nil {
		return nil, fmt.Errorf("Args not null.")
	} else if len(recs) != 2 {
		return nil, fmt.Errorf("Want 2 records, have %d", len(recs))
	}
	recsKeys := make([]string, 0, 2)
	recsValues := make([][]byte, 0, 2)
	for k, v := range recs {
		recsKeys = append(recsKeys, k)
		recsValues = append(recsValues, v)
	}
	out := make(map[string][]byte, 2)
	out[recsKeys[0]] = recsValues[1]
	out[recsKeys[1]] = recsValues[0]
	return out, nil
}
