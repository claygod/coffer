package coffer

// Coffer
// API tests
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	// "github.com/claygod/coffer/domain"
	// "github.com/claygod/coffer/services"
	// "github.com/claygod/coffer/services/filenamer"
	// "github.com/claygod/coffer/services/journal"
	// "github.com/claygod/coffer/services/repositories/handlers"
	// "github.com/claygod/coffer/services/repositories/records"
	// "github.com/claygod/coffer/reports"
	"github.com/claygod/coffer/reports/codes"
	"github.com/claygod/coffer/services/journal"
	"github.com/claygod/coffer/services/resources"

	// "github.com/claygod/coffer/services/startstop"
	"github.com/claygod/coffer/usecases"
	// "github.com/claygod/tools/logger"
	// "github.com/claygod/tools/porter"
)

const dirPath string = "./test/"

func TestCofferWriteListReadList(t *testing.T) {
	defer forTestClearDir(dirPath)
	cof1, err, wrn := createNewCoffer()
	if err != nil {
		t.Error(err)
		return
	} else if wrn != nil {
		t.Log(wrn)
	}
	if !cof1.Start() {
		t.Errorf("Failed to start (cof)")
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
	cof1, err, wrn := createNewCofferLength4(3, 7)
	if err != nil {
		t.Error(err)
		return
	} else if wrn != nil {
		t.Log(wrn)
	}
	if !cof1.Start() {
		t.Errorf("Failed to start (cof)")
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
	cof1, err, wrn := createNewCoffer()
	if err != nil {
		t.Error(err)
		return
	} else if wrn != nil {
		t.Log(wrn)
	}
	if !cof1.Start() {
		t.Errorf("Failed to start (cof)")
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
	t.Log("Stage1")
	cof1, err, wrn := createNewCoffer()
	if err != nil {
		t.Error(err)
		return
	} else if wrn != nil {
		t.Log(wrn)
	}
	if !cof1.Start() {
		t.Errorf("Failed to start (cof)")
		return
	}
	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb")); rep.Code > codes.Warning || rep.Error != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if rep := cof1.Count(); rep.Count != 9 {
		t.Errorf("Records (cof1) count, have %d, wand 9.", rep.Count)
		return
	}
	cof1.Stop()
	b1, err := ioutil.ReadFile(dirPath + "4.log") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ioutil.ReadFile(dirPath + "5.checkpoint") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	os.Remove(dirPath + "2.checkpoint")
	os.Remove(dirPath + "5.checkpoint")
	//time.Sleep(5000 * time.Millisecond)
	// специально портим один файл, и одна запись в нём при скачке должна быть потеряна
	t.Log("Stage2")
	if err := ioutil.WriteFile(dirPath+"4.log", b1[:len(b1)-2], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	// t.Log("Stage21")
	cof2, err, wrn := createNewCoffer()
	if err != nil {
		t.Log(err)
		return
	} else if wrn != nil {
		t.Log(wrn)
	}

	if !cof2.Start() {
		t.Errorf("Failed to start (cof2)")
		return
	}

	if rep := cof2.Count(); rep.Count != 8 { // одна запись поломана и её нет, а почему-то скачена
		t.Errorf("Records (cof2) count, have %d, wand 8.", rep.Count)
		return
	}
	os.Remove(dirPath + "5.log")
	os.Remove(dirPath + "6.checkpoint")
	time.Sleep(5000 * time.Millisecond)
	// // переименовываем один файл, в результате получив нормальный после битого
	// // но этот последний файл не должен быть загружен, т.к. загрузка должна остановиться на нём
	t.Log("Stage3")
	os.Rename(dirPath+"3.log", dirPath+"5.log")
	_, err, wrn = createNewCoffer()
	if err == nil {
		t.Error("Want error (The spoiled log...)")
		return
	} else {
		t.Log(wrn)
	}
}

func TestCofferLoadFromCheckpoint(t *testing.T) {
	defer forTestClearDir(dirPath)
	cof1, err, wrn := createNewCoffer()
	if err != nil {
		t.Error(err)
		return
	} else if wrn != nil {
		t.Log(wrn)
	}
	if !cof1.Start() {
		t.Errorf("Failed to start (cof)")
		return
	}
	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb")); rep.Code > codes.Warning || rep.Error != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	cof1.Stop()
	b1, err := ioutil.ReadFile(dirPath + "5.checkpoint") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	forTestClearDir(dirPath)
	// проверяем загрузку с нормального, небитого файла
	t.Log("Stage1")
	if err := ioutil.WriteFile(dirPath+"5.checkpoint", b1, os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err, wrn := createNewCoffer()
	if err != nil {
		t.Log(err)
		return
	} else if wrn != nil {
		t.Log(wrn)
	}

	if !cof2.Start() {
		t.Errorf("Failed to start (cof2)")
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
	if err := ioutil.WriteFile(dirPath+"5.checkpoint", b1[:len(b1)-2], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof3, err, wrn := createNewCoffer()
	if err != nil {
		t.Log(err)
		return
	} else if wrn != nil {
		t.Log(wrn)
	}

	if !cof3.Start() {
		t.Errorf("Failed to start (cof3)")
		return
	}

	if rep := cof3.Count(); rep.Count != 0 { // все записи битого чекпоинта должны быть проигнорированы
		t.Errorf("Records (cof3) count, have %d, wand 0.", rep.Count)
		return
	}
}

func TestCofferLoadFromFalseCheckpointTrueLogs(t *testing.T) {
	defer forTestClearDir(dirPath)
	cof1, err, wrn := createNewCoffer()
	if err != nil {
		t.Error(err)
		return
	} else if wrn != nil {
		t.Log(wrn)
	}
	if !cof1.Start() {
		t.Errorf("Failed to start (cof)")
		return
	}
	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb")); rep.Code > codes.Warning || rep.Error != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	cof1.Stop()
	b1, err := ioutil.ReadFile(dirPath + "5.checkpoint") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	// проверяем загрузку с нормального, небитого файла
	t.Log("Stage1")
	if err := ioutil.WriteFile(dirPath+"5.checkpoint", b1[:len(b1)-2], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err, wrn := createNewCoffer()
	if err != nil {
		t.Log(err)
		return
	} else if wrn != nil {
		t.Log(wrn)
	}

	if !cof2.Start() {
		t.Errorf("Failed to start (cof2)")
		return
	}

	if rep := cof2.Count(); rep.Count != 9 { // не все записи скачены, хотя при битом чекпоинте всё должно было быть скачено с логов
		t.Errorf("Records (cof2) count, have %d, wand 9.", rep.Count)
		return
	}
	cof2.Stop()
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
	return New(cnf)
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
	return New(cnf)
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
