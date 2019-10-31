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
	//"github.com/claygod/coffer/services/journal"
	//"github.com/claygod/coffer/services/repositories/handlers"
	//"github.com/claygod/coffer/services/resources"
	// "github.com/claygod/coffer/services/startstop"
	//"github.com/claygod/coffer/usecases"
	// "github.com/claygod/tools/logger"
	// "github.com/claygod/tools/porter"
)

const dirPath string = "./test/"

func TestCofferCleanDir(t *testing.T) {
	forTestClearDir(dirPath)
}

func TestNewDirNotFound(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)

	if _, err, _ := Db("./not_found_dir/").BatchSize(2000).
		MaxRecsPerOperation(10).Create(); err == nil {
		t.Error("Want error, have nil.")
	}

}

func TestWriteList(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	defer cof1.Stop()
	cof1.Write("aaa", []byte("111"))
	if rep := cof1.WriteList(map[string][]byte{"aaa": []byte("1"), "bbb": []byte("2")}, true); !rep.IsCodeErrRecordsFound() {
		t.Errorf("Operation `WriteList`(1) results: code=%d ,  found=%v, err=%v.", rep.Code, rep.Found, rep.Error)
		return
	}
	if rep := cof1.WriteList(map[string][]byte{"ccc": []byte("1"), "ddd": []byte("2")}, true); rep.IsCodeErrRecordsFound() {
		t.Errorf("Operation `WriteList`(2) results: code=%d ,  found=%v, err=%v.", rep.Code, rep.Found, rep.Error)
		return
	}
	if rep := cof1.Count(); !rep.IsCodeOk() || rep.Error != nil || rep.Count != 3 {
		t.Errorf("Operation `Count`(3) results: code=%d , count=%v, err=%v.", rep.Code, rep.Count, rep.Error)
		return
	}
}

func TestDeleteList(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	defer cof1.Stop()
	cof1.Write("aaa", []byte("111"))
	cof1.Write("bbb", []byte("222"))
	cof1.Write("fff", []byte("333"))
	cof1.Write("ggg", []byte("333"))
	// -- deleteList
	if rep := cof1.deleteList([]string{"aaa"}, true); !rep.IsCodeOk() {
		t.Errorf("Operation `deleteList`(1) results: code=%d , removed=%v, not_found=%v, err=%v.", rep.Code, rep.Removed, rep.NotFound, rep.Error)
		return
	}
	if rep := cof1.deleteList([]string{"ccc"}, true); !rep.IsCodeErrNotFound() {
		t.Errorf("Operation `deleteList`(2) results: code=%d , removed=%v, not_found=%v, err=%v.", rep.Code, rep.Removed, rep.NotFound, rep.Error)
		return
	}
	if rep := cof1.deleteList([]string{"bbb", "ccc"}, false); !rep.IsCodeOk() || rep.Error != nil ||
		len(rep.Removed) != 1 || rep.Removed[0] != "bbb" ||
		len(rep.NotFound) != 1 || rep.NotFound[0] != "ccc" {
		t.Errorf("Operation `deleteList`(3) results: code=%d , removed=%v, not_found=%v, err=%v.", rep.Code, rep.Removed, rep.NotFound, rep.Error)
		return
	}
	cof1.hasp.Stop()
	if rep := cof1.deleteList([]string{"xxx"}, true); !rep.IsCodePanicStopped() {
		t.Errorf("Operation `deleteList`(4) results: code=%d , removed=%v, not_found=%v, err=%v.", rep.Code, rep.Removed, rep.NotFound, rep.Error)
		return
	}
	cof1.hasp.Start()

	// -- DeleteList
	if rep := cof1.DeleteListStrict([]string{"fff"}); !rep.IsCodeOk() {
		t.Errorf("Operation `deleteList`(1) results: code=%d , removed=%v, not_found=%v, err=%v.", rep.Code, rep.Removed, rep.NotFound, rep.Error)
		return
	}
	if rep := cof1.DeleteListStrict([]string{"ccc"}); !rep.IsCodeErrNotFound() {
		t.Errorf("Operation `deleteList`(2) results: code=%d , removed=%v, not_found=%v, err=%v.", rep.Code, rep.Removed, rep.NotFound, rep.Error)
		return
	}
	if rep := cof1.DeleteListOptional([]string{"ggg", "ccc"}); !rep.IsCodeOk() || rep.Error != nil ||
		len(rep.Removed) != 1 || rep.Removed[0] != "ggg" ||
		len(rep.NotFound) != 1 || rep.NotFound[0] != "ccc" {
		t.Errorf("Operation `deleteList`(3) results: code=%d , removed=%v, not_found=%v, err=%v.", rep.Code, rep.Removed, rep.NotFound, rep.Error)
		return
	}
	cof1.hasp.Stop()
	if rep := cof1.DeleteListStrict([]string{"xxx"}); !rep.IsCodePanicStopped() {
		t.Errorf("Operation `deleteList`(4) results: code=%d , removed=%v, not_found=%v, err=%v.", rep.Code, rep.Removed, rep.NotFound, rep.Error)
		return
	}
	cof1.hasp.Start()
}

func TestCount(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	defer cof1.Stop()
	cof1.Write("aaa", []byte("111"))
	cof1.Write("bbb", []byte("222"))

	if rep := cof1.Count(); !rep.IsCodeOk() || rep.Error != nil || rep.Count != 2 {
		t.Errorf("Operation `Count`(1) results: code=%d , count=%v, err=%v.", rep.Code, rep.Count, rep.Error)
		return
	}

	cof1.hasp.Stop()
	if rep := cof1.Count(); !rep.IsCodePanicStopped() || rep.Error == nil || rep.Count != 0 {
		t.Errorf("Operation `Count`(2) results: code=%d , count=%v, err=%v.", rep.Code, rep.Count, rep.Error)
		return
	}
	//cof1.hasp.Start()
}

func TestCountUnsafe(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	defer cof1.Stop()
	cof1.Write("aaa", []byte("111"))
	cof1.Write("bbb", []byte("222"))

	if rep := cof1.CountUnsafe(); !rep.IsCodeOk() || rep.Error != nil || rep.Count != 2 {
		t.Errorf("Operation `Count`(1) results: code=%d , count=%v, err=%v.", rep.Code, rep.Count, rep.Error)
		return
	}

	cof1.hasp.Stop()
	if rep := cof1.CountUnsafe(); rep.IsCodePanicStopped() || rep.Error != nil || rep.Count != 2 {
		t.Errorf("Operation `Count`(2) results: code=%d , count=%v, err=%v.", rep.Code, rep.Count, rep.Error)
		return
	}
	//cof1.hasp.Start()
}

func TestCofferReadListPrefixSuffix(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	defer cof1.Stop()
	cof1.WriteList(map[string][]byte{"pr1-suf1": {1}, "pr1-suf2": {1}, "pr2-suf1": {1}, "pr2-suf2": {1}, "pr3-suf2": {1}}, false)
	if rep := cof1.RecordsListWithPrefix("pr1"); !rep.IsCodeOk() || rep.Error != nil || len(rep.Data) != 2 {
		t.Errorf("Operation `RecordsListWithPrefix`(1) results: code=%d , data=%v, err=%v.", rep.Code, rep.Data, rep.Error)
	}
	if rep := cof1.RecordsListWithSuffix("suf2"); !rep.IsCodeOk() || rep.Error != nil || len(rep.Data) != 3 {
		t.Errorf("Operation `RecordsListWithPrefix`(2) results: code=%d , data=%v, err=%v.", rep.Code, rep.Data, rep.Error)
	}
	if rep := cof1.RecordsListWithSuffix("suf7"); !rep.IsCodeOk() || rep.Error != nil || len(rep.Data) != 0 {
		t.Errorf("Operation `RecordsListWithPrefix`(3) results: code=%d , data=%v, err=%v.", rep.Code, rep.Data, rep.Error)
	}
	cof1.hasp.Stop()
	if rep := cof1.RecordsListWithPrefix("pr1"); rep.IsCodeOk() || rep.Error == nil {
		t.Errorf("Operation `RecordsListWithPrefix`(5) results: code=%d , data=%v, err=%v.", rep.Code, rep.Data, rep.Error)
	}
	if rep := cof1.RecordsListWithSuffix("suf2"); rep.IsCodeOk() || rep.Error == nil {
		t.Errorf("Operation `RecordsListWithPrefix`(2) results: code=%d , data=%v, err=%v.", rep.Code, rep.Data, rep.Error)
	}
}

func TestCofferReadListUnsafe(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	defer cof1.Stop()
	cof1.WriteList(map[string][]byte{"pr1-suf1": {1}, "pr1-suf2": {1}, "pr2-suf1": {1}, "pr2-suf2": {1}, "pr3-suf2": {1}}, false)

	if rep := cof1.RecordsListUnsafe(); !rep.IsCodeOk() || rep.Error != nil || len(rep.Data) != 5 {
		t.Errorf("RecordsListUnsafe`(1) results: code=%d , data=%v, err=%v.", rep.Code, rep.Data, rep.Error)
	}
	if rep := cof1.ReadListUnsafe([]string{"pr1-suf1", "pr2-suf2"}); !rep.IsCodeOk() || rep.Error != nil || len(rep.Data) != 2 {
		t.Errorf("ReadListUnsafe`(2) results: code=%d , data=%v, err=%v.", rep.Code, rep.Data, rep.Error)
	}
	//longKey := "loooonnggggg"
	longKey := make([]byte, 2000)
	for i := 0; i < 2000; i++ {
		longKey[i] = 101
	}

	if rep := cof1.ReadListUnsafe([]string{string(longKey), "pr2-suf2"}); !rep.IsCodeErrExceedingMaxKeyLength() || rep.Error == nil {
		t.Errorf("ReadListUnsafe`(2) results: code=%d , data=%v, err=%v.", rep.Code, rep.Data, rep.Error)
	}
	cof1.hasp.Stop()
	if rep := cof1.RecordsListUnsafe(); !rep.IsCodeOk() || rep.Error != nil || len(rep.Data) != 5 {
		t.Errorf("RecordsListUnsafe`(3) results: code=%d , data=%v, err=%v.", rep.Code, rep.Data, rep.Error)
	}
	if rep := cof1.ReadListUnsafe([]string{"pr1-suf1", "pr2-suf2"}); !rep.IsCodeOk() || rep.Error != nil || len(rep.Data) != 2 {
		t.Errorf("ReadListUnsafe`(4) results: code=%d , data=%v, err=%v.", rep.Code, rep.Data, rep.Error)
	}
	// cof1.hasp.Start()
}

func TestCofferRecordsList(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	defer cof1.Stop()
	cof1.WriteList(map[string][]byte{"pr1-suf1": {1}, "pr1-suf2": {1}, "pr2-suf1": {1}, "pr2-suf2": {1}, "pr3-suf2": {1}}, false)

	if rep := cof1.RecordsList(); !rep.IsCodeOk() || rep.Error != nil || len(rep.Data) != 5 {
		t.Errorf("RecordsList`(1) results: code=%d , data=%v, err=%v.", rep.Code, rep.Data, rep.Error)
	}
	cof1.hasp.Stop()
	if rep := cof1.RecordsList(); rep.IsCodeOk() || rep.Error == nil {
		t.Errorf("RecordsList`(2) results: code=%d , data=%v, err=%v.", rep.Code, rep.Data, rep.Error)
	}
	// cof1.hasp.Start()
}

func TestCofferTransaction(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	defer cof1.Stop()

	cof1.Write("aaa", []byte("111"))
	cof1.Write("bbb", []byte("222"))
	if rep := cof1.Transaction("exchange", []string{"aaa", "bbb"}, nil); rep.IsCodeError() {
		t.Error(err)
		return
	} else if rep.Data == nil {
		t.Error("Want notnull")
		return
	} else if bt, ok := rep.Data["aaa"]; !ok || string(bt) != "222" {
		t.Errorf("Want aaa==222 , have %v", rep.Data)
		return
	} else if bt, ok := rep.Data["bbb"]; !ok || string(bt) != "111" {
		t.Errorf("Want bbb==111 , have %v", rep.Data)
		return
	}
	// the number of records should not change
	if rep := cof1.Count(); rep.Count != 2 {
		t.Errorf("Records (cof1) count, have %d, want 2.", rep.Count)
		return
	}
	// ---
	rep := cof1.ReadList([]string{"aaa", "bbb"})
	if rep.IsCodeError() {
		t.Errorf("Transaction results: code=%d , data=%v, not_found=%v, err=%v.", rep.Code, rep.Data, rep.NotFound, rep.Error)
		return
	} else if string(rep.Data["aaa"]) != "222" || string(rep.Data["bbb"]) != "111" {
		t.Errorf("Want aaa==222 bbb==111 , have aaa=%s bbb==%s ", string(rep.Data["aaa"]), string(rep.Data["bbb"]))
	}
	cof1.Stop()
	if rep := cof1.Transaction("exchange", []string{"aaa", "bbb"}, nil); !rep.IsCodePanicStopped() {
		t.Errorf("Have code `PanicStopped` want `%d` ", rep.Code)
		//return
	}
}

func TestCofferTransactionChain(t *testing.T) {
	forTestClearDir(dirPath)
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	defer cof1.Stop()

	cof1.Write("aaa", []byte("111"))
	cof1.Write("bbb", []byte("222"))
	cof1.Write("ccc", []byte("333"))
	cof1.Write("ddd", []byte("444"))

	if rep := cof1.Transaction("exchange", []string{"ddd", "ccc"}, nil); rep.IsCodeError() {
		t.Error(err)
		return
	}
	if rep := cof1.Transaction("exchange", []string{"ccc", "bbb"}, nil); rep.IsCodeError() {
		t.Error(err)
		return
	}
	if rep := cof1.Transaction("exchange", []string{"bbb", "aaa"}, nil); rep.IsCodeError() {
		t.Error(err)
		return
	}
	// the number of records should not change
	if rep := cof1.Count(); rep.Count != 4 {
		t.Errorf("Records (cof1) count, have %d, want 4.", rep.Count)
		return
	}
	// ---
	rep := cof1.ReadList([]string{"aaa", "bbb", "ccc", "ddd"})
	if rep.IsCodeError() {
		t.Errorf("Transaction results: code=%d , data=%v, not_found=%v, err=%v.", rep.Code, rep.Data, rep.NotFound, rep.Error)
		return
	} else if string(rep.Data["aaa"]) != "444" || string(rep.Data["bbb"]) != "111" || string(rep.Data["ccc"]) != "222" || string(rep.Data["ddd"]) != "333" {
		fmt.Println(rep.Data)
		t.Errorf("Want aaa==444 bbb==111 ccc==222 ddd==333 , have aaa=%s bbb==%s ccc==%s ddd==%s",
			string(rep.Data["aaa"]), string(rep.Data["bbb"]), string(rep.Data["ccc"]), string(rep.Data["ddd"]))
	}
	cof1.Stop()
	if rep := cof1.Transaction("exchange", []string{"aaa", "bbb"}, nil); !rep.IsCodePanicStopped() {
		t.Errorf("Have code `PanicStopped` want `%d` ", rep.Code)
		//return
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
	if !cof1.Start() {
		t.Error("Could not start the application (1)")
		return
	}
	defer cof1.Stop()

	cof1.Write("aaa", []byte("111"))
	cof1.Write("bbb", []byte("222"))
	if rep := cof1.Transaction("exchange", []string{"xxx", "yyy"}, nil); !rep.IsCodeErrReadRecords() {
		t.Errorf("Want codes.ErrReadRecords , have %v", rep.Code)
		t.Error(rep.Error)
		return
	}
}

func TestCofferTransactionRecordsZeroLenKeys(t *testing.T) {
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
	if !cof1.Start() {
		t.Error("Could not start the application (1)")
		return
	}
	defer cof1.Stop()

	if rep := cof1.Transaction("exchange", []string{""}, nil); !rep.IsCodeErrExceedingZeroKeyLength() {
		t.Errorf("Want codes.ErrExceedingZeroKeyLength , have %v", rep.Code)
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
	if !cof1.Start() {
		t.Error("Could not start the application (1)")
		return
	}
	defer cof1.Stop()

	if rep := cof1.Transaction("exchange", []string{"xxxxx"}, nil); !rep.IsCodeErrExceedingMaxKeyLength() {
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
	if !cof1.Start() {
		t.Error("Could not start the application (1)")
		return
	}
	defer cof1.Stop()

	if rep := cof1.Transaction("exchange", []string{"x", "x", "x", "x", "x", "x", "x",
		"x", "x", "x", "x", "x", "x", "x"}, nil); !rep.IsCodeErrRecordLimitExceeded() {
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
	}
	defer cof1.Stop()

	if rep := cof1.Transaction("exchangeXXX", []string{"aaa", "bbb"}, nil); !rep.IsCodeErrHandlerNotFound() {
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
	// if err := cof1.Save(); err != nil {
	// 	t.Error(err)
	// }
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
	// if err := cof1.Save(); err != nil {
	// 	t.Error(err)
	// }

	if rep := cof1.Count(); !rep.IsCodePanicStopped() {
		//t.Errorf("Report: %v", rep)
	}
}

// func TestCofferStopHard(t *testing.T) {
// 	forTestClearDir(dirPath)
// 	cof1, err := createAndStartNewCofferLength(t, 4, 7)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	if !cof1.Start() {
// 		t.Error("Could not start the application (1)")
// 	}
// 	if err := cof1.StopHard(); err != nil {
// 		t.Error(err)
// 	}
// }

func TestCofferWriteRead(t *testing.T) {
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferLength(t, 4, 7)
	if err != nil {
		t.Error(err)
		return
	}
	// без ошибок
	t.Log("Stage1")
	if rep := cof1.Write("aa", []byte("bbb")); !rep.IsCodeOk() || rep.Error != nil {
		t.Error("Write: want code 0 (Ok), have code: ", rep.Code, " Resp. err.: ", rep.Error)
		return
	}
	if rep := cof1.Read("aa"); !rep.IsCodeOk() || rep.Error != nil || rep.Data == nil {
		t.Error("Read: want code 0 (Ok), have code: ", rep.Code, " Resp. err.: ", rep.Error, " Resp. data: ", rep.Data)
		return
	}
	// -- too big key
	t.Log("Stage2")
	if rep := cof1.Write("aaaaa", []byte("bbb")); !rep.IsCodeErrExceedingMaxKeyLength() || rep.Error == nil {
		t.Error("Write: want code `ErrExceedingMaxKeyLength`, have code: ", rep.Code, " Resp. err.: ", rep.Error)
		return
	}
	// -- too important
	t.Log("Stage3")
	if rep := cof1.Write("dd", []byte("cccccccccccc")); !rep.IsCodeErrExceedingMaxValueSize() || rep.Error == nil {
		t.Error("Write: want code `ErrExceedingMaxValueSize`, have code: ", rep.Code, " Resp. err.: ", rep.Error)
		return
	}
	// -- trying to read a nonexistent record
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
	// -- write down the list
	t.Log("Stage1")
	if rep := cof1.WriteList(req, false); rep.IsCodeError() || rep.Error != nil {
		t.Error(err)
	}
	if rep := cof1.Count(); rep.Count != 9 {
		t.Errorf("Records (cof1) count, have %d, wand 9.", rep.Count)
		return
	}
	//longKey := "loooonnggggg"
	longKey := make([]byte, 2000)
	for i := 0; i < 2000; i++ {
		longKey[i] = 101
	}
	//longValue := "loooonnggggg"
	longValue := make([]byte, 20000)
	for i := 0; i < 20000; i++ {
		longValue[i] = 102
	}
	if rep := cof1.WriteList(map[string][]byte{string(longKey): []byte("zzz")}, false); !rep.IsCodeErrExceedingMaxKeyLength() || rep.Error == nil {
		t.Error(rep)
		return
	}
	if rep := cof1.WriteList(map[string][]byte{"shortKey": longValue}, false); !rep.IsCodeErrExceedingMaxValueSize() || rep.Error == nil {
		t.Error(rep)
		return
	}
	// -- read the real list
	t.Log("Stage2")
	rep := cof1.ReadList([]string{"aasa10", "aasa11"})
	if !rep.IsCodeOk() || rep.Error != nil || len(rep.Data) != 2 || len(rep.NotFound) != 0 {
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
	// -- trying to count non-existent entries too
	t.Log("Stage3")
	rep = cof1.ReadList([]string{"aasa10", "aasa90"})
	if !rep.IsCodeErrReadRecords() || rep.Error != nil || len(rep.Data) != 1 || len(rep.NotFound) != 1 {
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

	if rep := cof1.ReadList([]string{"aasa10", "aasa90"}); !rep.IsCodePanicStopped() {
		t.Errorf("Want cote `stopped` have code  `%d` .", rep.Code)
	}
	cof1.Stop()
	if rep := cof1.Transaction("exchange", []string{"aaa", "bbb"}, nil); !rep.IsCodePanicStopped() {
		t.Errorf("Have code `PanicStopped` want `%d` ", rep.Code)
		//return
	}

	if rep := cof1.WriteList(req, false); !rep.IsCodePanicStopped() || rep.Error == nil {
		t.Errorf("Have code `PanicStopped` want `%d` ", rep.Code)
	}
}

func TestCofferWriteListUnsafeReadList(t *testing.T) {
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
	// -- write down the list
	t.Log("Stage1")
	if rep := cof1.WriteListUnsafe(req); rep.IsCodeError() || rep.Error != nil {
		t.Error(err)
		return
	}
	if rep := cof1.Count(); rep.Count != 9 {
		t.Errorf("Records (cof1) count, have %d, wand 9.", rep.Count)
		return
	}
	//longKey := "loooonnggggg"
	longKey := make([]byte, 2000)
	for i := 0; i < 2000; i++ {
		longKey[i] = 101
	}
	//longValue := "loooonnggggg"
	longValue := make([]byte, 20000)
	for i := 0; i < 20000; i++ {
		longValue[i] = 102
	}
	if rep := cof1.WriteListUnsafe(map[string][]byte{string(longKey): []byte("zzz")}); !rep.IsCodeErrExceedingMaxKeyLength() || rep.Error == nil {
		t.Error(rep)
		return
	}
	if rep := cof1.WriteListUnsafe(map[string][]byte{"shortKey": longValue}); !rep.IsCodeErrExceedingMaxValueSize() || rep.Error == nil {
		t.Error(rep)
		return
	}
	// -- read the real list
	t.Log("Stage2")
	rep := cof1.ReadList([]string{"aasa10", "aasa11"})
	if !rep.IsCodeOk() || rep.Error != nil || len(rep.Data) != 2 || len(rep.NotFound) != 0 {
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
	// -- trying to count non-existent entries too
	t.Log("Stage3")
	rep = cof1.ReadList([]string{"aasa10", "aasa90"})
	if !rep.IsCodeErrReadRecords() || rep.Error != nil || len(rep.Data) != 1 || len(rep.NotFound) != 1 {
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

	if rep := cof1.ReadList([]string{"aasa10", "aasa90"}); !rep.IsCodePanicStopped() {
		t.Errorf("Want code `stopped` have code  `%d` .", rep.Code)
	}
	cof1.Stop()
	if rep := cof1.Transaction("exchange", []string{"aaa", "bbb"}, nil); !rep.IsCodePanicStopped() {
		t.Errorf("Want code `PanicStopped` have `%d` ", rep.Code)
		//return
	}

	if rep := cof1.WriteListUnsafe(req); !rep.IsCodePanicWAL() || rep.Error == nil {
		t.Errorf("Want code `PanicWAL` have `%d` ", rep.Code)
	}
}

func TestCofferKeyLength(t *testing.T) {
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCofferLength(t, 3, 7)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Stage1")
	if rep := cof1.Write("aa1aa", []byte("bbsb10")); !rep.IsCodeErrExceedingMaxKeyLength() || rep.Error == nil {
		t.Error(codes.ErrExceedingMaxKeyLength, rep, err)
		return
	} else if rep := cof1.Write("aa1", []byte("bb1bbbbbbbbbbbbb")); !rep.IsCodeErrExceedingMaxValueSize() || rep.Error == nil {
		t.Error(codes.ErrExceedingMaxValueSize, rep, err)
		return
	}
	t.Log("Stage2")
	if rep := cof1.Write("aa1", []byte("bb1")); !rep.IsCodeOk() || rep.Error != nil {
		t.Error(codes.Ok, rep, err)
		return
	} else if rep := cof1.Read("aa1"); !rep.IsCodeOk() || rep.Error != nil {
		t.Error(rep, err)
		return
	} else if rep := cof1.Read("aa1aa"); !rep.IsCodeErrExceedingMaxKeyLength() || rep.Error == nil {
		t.Error(codes.ErrExceedingMaxKeyLength, rep, err)
		return
	}
	t.Log("Stage3")
	if rep := cof1.Delete("aa1aa"); !rep.IsCodeErrExceedingMaxKeyLength() || rep.Error == nil {
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
	// trying to record too many records at a time
	t.Log("Stage1")
	reqWriteList := make(map[string][]byte)
	reqReadList := make([]string, 0)
	for i := 10; i < 22; i++ {
		reqWriteList["aasa"+strconv.Itoa(i)] = []byte("bbsb" + strconv.Itoa(i))
		reqReadList = append(reqReadList, "aasa"+strconv.Itoa(i))
	}
	rep := cof1.WriteList(reqWriteList, false)
	if !rep.IsCodeErrRecordLimitExceeded() || rep.Error == nil {
		t.Errorf("Want `ErrRecordLimitExceeded` have code `%d`", rep.Code)
		t.Error(rep.Error)
		return
	}
	t.Log(rep)
	// trying to read too many records at one time
	t.Log("Stage2")
	rep2 := cof1.ReadList(reqReadList)
	if !rep2.IsCodeErrRecordLimitExceeded() || rep2.Error == nil {
		t.Errorf("Want `ErrRecordLimitExceeded` have code `%d`", rep2.Code)
		t.Error(rep2.Error)
		return
	}
	// attempt to delete Strict at once too many records
	t.Log("Stage3")
	rep3 := cof1.DeleteListStrict(reqReadList)
	if !rep3.IsCodeErrRecordLimitExceeded() || rep3.Error == nil {
		t.Errorf("Want `ErrRecordLimitExceeded` have code `%d`", rep2.Code)
		t.Error(rep3.Error)
		return
	}
	// attempt to delete Optional at one time too many records
	t.Log("Stage3")
	rep3 = cof1.DeleteListOptional(reqReadList)
	if !rep3.IsCodeErrRecordLimitExceeded() || rep3.Error == nil {
		t.Errorf("Want `ErrRecordLimitExceeded` have code `%d`", rep2.Code)
		t.Error(rep3.Error)
		return
	}
}

func TestCofferLoadFromLogs(t *testing.T) {
	defer forTestClearDir(dirPath)

	// fill the base
	t.Log("Stage1")
	cof1, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}
	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb")); rep.IsCodeError() || rep.Error != nil {
			t.Error(err)
			return
		}
		//time.Sleep(100 * time.Millisecond)
	}
	if rep := cof1.Count(); rep.Count != 9 {
		t.Errorf("Records (cof1) count, have %d, wand 9.", rep.Count)
		return
	}
	cof1.Stop()
	b1, err := ioutil.ReadFile(dirPath + "1000000001.log") // save to memory
	if err != nil {
		t.Error(err)
		return
	}
	b2, err := ioutil.ReadFile(dirPath + "1000000002.log") // save to memory
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
	}
	t.Log("Load true logs OK")

	cof111.Stop()
	forTestClearDir(dirPath)

	// specially we corrupt one file, and one record in it should be lost when downloading
	t.Log("Stage3")
	if err := ioutil.WriteFile(dirPath+"1000000003.log", b1, os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	if err := ioutil.WriteFile(dirPath+"1000000004.log", b2[:len(b2)-1], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}
	if rep := cof2.Count(); rep.Count != 8 { // one record is broken and not there, but for some reason downloaded
		t.Errorf("Records (cof2) count, have %d, want 8.", rep.Count)
		return
	}
	t.Log("Load false logs OK")

	cof2.Stop()
	forTestClearDir(dirPath)

	// rename one file, as a result, getting normal after the beaten
	// but this last file should not be uploaded because loading should stop on it
	t.Log("Stage4")
	if err := ioutil.WriteFile(dirPath+"1000000009.log", b1, os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	if err := ioutil.WriteFile(dirPath+"1000000008.log", b2[:len(b2)-1], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	_, err, wrn := createNewCoffer()
	if err == nil {
		t.Error("Want error (The spoiled log...)")
		return
	}
	t.Log(wrn)
	t.Log("Load false/true logs OK")

	//time.Sleep(15000 * time.Millisecond)
}

func TestCofferLoadFromLogsTransaction(t *testing.T) {
	defer forTestClearDir(dirPath)

	// fill the base
	t.Log("Stage1")
	cof1, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}

	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb"+strconv.Itoa(i))); rep.IsCodeError() || rep.Error != nil {
			t.Error(err)
			return
		}
	}
	if rep := cof1.Transaction("exchange", []string{"aasa10", "aasa11"}, nil); rep.IsCodeError() {
		t.Error(rep)
		return
	}

	if rep := cof1.Count(); rep.Count != 9 {
		t.Errorf("Records (cof1) count, have %d, wand 9.", rep.Count)
		return
	}
	cof1.Stop()
	b1, err := ioutil.ReadFile(dirPath + "1000000001.log")
	if err != nil {
		t.Error(err)
		return
	}
	b2, err := ioutil.ReadFile(dirPath + "1000000002.log")
	if err != nil {
		t.Error(err)
		return
	}

	// try to boot from the logs
	t.Log("Stage2")
	cof111, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	if rep := cof111.Count(); rep.Count != 9 {
		t.Errorf("Records (cof111) count, have %d, wand 9.", rep.Count)
		return
	}
	t.Log("Load true logs OK")
	if rep := cof111.Read("aasa10"); string(rep.Data) != "bbsb11" {
		t.Errorf("Record have %s, wand `bbsb11`.", string(rep.Data))
		return
	}
	cof111.Stop()
	forTestClearDir(dirPath)

	// specially we corrupt one file, and one record in it should be lost when downloading
	t.Log("Stage3")
	if err := ioutil.WriteFile(dirPath+"1000000003.log", b1, os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	if err := ioutil.WriteFile(dirPath+"1000000004.log", b2[:len(b2)-1], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}
	if rep := cof2.Count(); rep.Count != 9 { // one record is broken and not there, but for some reason downloaded
		t.Errorf("Records (cof2) count, have %d, want 9.", rep.Count)
		return
	}
	t.Log("Load false logs OK")

	if rep := cof2.Read("aasa10"); string(rep.Data) != "bbsb10" { // now the last action (transaction) is canceled
		t.Errorf("Record have %s, wand `bbsb10`.", string(rep.Data))
		return
	}
	cof2.Stop()
	forTestClearDir(dirPath)

	// rename a file, the result will be normal after broken
	// but this last file should not be uploaded because loading should stop on it
	t.Log("Stage4")
	if err := ioutil.WriteFile(dirPath+"1000000009.log", b1, os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	if err := ioutil.WriteFile(dirPath+"1000000008.log", b2[:len(b2)-1], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	_, err, wrn := createNewCofferT()
	if err == nil {
		t.Error("Want error (The spoiled log...)")
		return
	}
	t.Log(wrn)
	t.Log("Load false/true logs OK")

}

func TestCofferLoadFromCheckpoint(t *testing.T) {
	defer forTestClearDir(dirPath)
	cof1, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}
	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb")); rep.IsCodeError() || rep.Error != nil {
			t.Error(err)
			return
		}
	}
	cof1.Stop()
	b1, err := ioutil.ReadFile(dirPath + "1000000003.checkpoint")
	if err != nil {
		t.Error(err)
		return
	}
	forTestClearDir(dirPath)
	// check the download from a normal file
	t.Log("Stage1")
	if err := ioutil.WriteFile(dirPath+"1000000003.checkpoint", b1, os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}

	if rep := cof2.Count(); rep.Count != 9 { // not all entries downloaded
		t.Errorf("Records (cof2) count, have %d, wand 9.", rep.Count)
		return
	}
	cof2.Stop()
	forTestClearDir(dirPath)
	// check download from a broken file
	t.Log("Stage2")
	if err := ioutil.WriteFile(dirPath+"1000000003.checkpoint", b1[:len(b1)-2], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof3, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}

	if rep := cof3.Count(); rep.Count != 0 { // all records of the broken checkpoint should be ignored
		t.Errorf("Records (cof3) count, have %d, wand 0.", rep.Count)
		return
	}
}

func TestCofferLoadFromCheckpointTransaction(t *testing.T) {
	defer forTestClearDir(dirPath)
	cof1, err, wrn := createNewCofferT()
	if err != nil {
		t.Error(err)
		return
	} else if wrn != nil {
		t.Error(wrn)
		return
	}
	cof1.Start()
	for i := 10; i < 19; i++ {
		rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb"+strconv.Itoa(i)))
		if rep.IsCodeError() || rep.Error != nil {
			t.Error(err)
			return
		}
	}
	if rep := cof1.Transaction("exchange", []string{"aasa10", "aasa11"}, nil); rep.IsCodeError() {
		t.Error(rep)
		return
	}
	cof1.Stop()
	time.Sleep(5 * time.Millisecond)
	b1, err := ioutil.ReadFile(dirPath + "1000000003.checkpoint") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	forTestClearDir(dirPath)
	// check the download from a normal file
	t.Log("Stage1")
	if err := ioutil.WriteFile(dirPath+"1000000003.checkpoint", b1, os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}

	if rep := cof2.Count(); rep.Count != 9 { // not all entries downloaded
		t.Errorf("Records (cof2) count, have %d, wand 9.", rep.Count)
		return
	}
	rep := cof2.Read("aasa10")
	if !rep.IsCodeOk() || rep.Error != nil {
		t.Errorf("Read error. Code: %d, data: %v, err: %v .", rep.Code, rep.Data, rep.Error)
		return
	}
	if string(rep.Data) != "bbsb11" {
		t.Errorf("Record want `bbsb11` have `%s`", string(rep.Data))
		return
	}
	cof2.Stop()
	forTestClearDir(dirPath)
	// check download from a broken file
	t.Log("Stage2")
	if err := ioutil.WriteFile(dirPath+"1000000003.checkpoint", b1[:len(b1)-2], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof3, err := createAndStartNewCoffer(t)
	if err != nil { // when loading from a damaged file, it is ignored
		t.Error(err)
		return
	}

	if rep := cof3.Count(); rep.Count != 0 { // all records of the broken checkpoint should be ignored
		t.Errorf("Records (cof3) count, have %d, wand 0.", rep.Count)
		return
	}
}

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
	cof1.Start()
	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb"+strconv.Itoa(i))); rep.IsCodeError() || rep.Error != nil {
			t.Error(err)
		}
	}
	cof1.Stop()
	b1, err := ioutil.ReadFile(dirPath + "1000000003.checkpoint")
	if err != nil {
		t.Error(err)
		return
	}
	// check the download from a normal file
	t.Log("Stage1")
	if err := ioutil.WriteFile(dirPath+"1000000003.checkpoint", b1[:len(b1)-2], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err := createAndStartNewCoffer(t)
	if err != nil {
		t.Error(err)
		return
	}

	if rep := cof2.Count(); rep.Count != 9 { // not all records are downloaded, although with a damaged checkpoint everything should have been downloaded from the logs
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
	for i := 10; i < 19; i++ {
		if rep := cof1.Write("aasa"+strconv.Itoa(i), []byte("bbsb"+strconv.Itoa(i))); rep.IsCodeError() || rep.Error != nil {
			t.Error(err)
		}
	}
	if rep := cof1.Transaction("exchange", []string{"aasa10", "aasa11"}, nil); rep.IsCodeError() {
		t.Error(rep)
		return
	}
	cof1.Stop()
	b1, err := ioutil.ReadFile(dirPath + "1000000003.checkpoint") // сохраняем в память
	if err != nil {
		t.Error(err)
		return
	}
	// check the download from a normal file
	t.Log("Stage1")
	if err := ioutil.WriteFile(dirPath+"1000000003.checkpoint", b1[:len(b1)-2], os.ModePerm); err != nil {
		t.Error(err)
		return
	}
	cof2, err := createAndStartNewCofferT(t)
	if err != nil {
		t.Error(err)
		return
	}

	if rep := cof2.Count(); rep.Count != 9 { // not all records are downloaded, although with a broken checkpoint everything should have been downloaded from the logs
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
		return nil, err
	} else if wrn != nil {
		t.Log(wrn)
	}
	if !cof1.Start() {
		return nil, fmt.Errorf("Failed to start (cof)")
	}
	return cof1, nil
}

func createAndStartNewCofferT(t *testing.T) (*Coffer, error) {
	cof1, err, wrn := createNewCofferT()
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
	return Db(dirPath).BatchSize(2000).
		LimitRecordsPerLogfile(5).
		FollowPause(400 * time.Millisecond).
		LogsByCheckpoint(2).
		AllowStartupErrLoadLogs(defaultAllowStartupErrLoadLogs). //--
		MaxKeyLength(defaultMaxKeyLength).                       //--
		MaxValueLength(defaultMaxValueLength).                   //-
		RemoveUnlessLogs(defaultRemoveUnlessLogs).               //--
		LimitMemory(int(defaultLimitMemory)).                    //--
		LimitDisk(int(defaultLimitDisk)).                        //--
		MaxRecsPerOperation(10).
		Create()

}

func createNewCofferT() (*Coffer, error, error) {

	hdlExch := domain.Handler(handlerExchange)
	return Db(dirPath).BatchSize(2000).
		LimitRecordsPerLogfile(5).
		FollowPause(400*time.Millisecond).
		LogsByCheckpoint(2).
		MaxRecsPerOperation(10).
		Handler("exchange", &hdlExch).
		Create()
}

func createNewCofferLength4(maxKeyLength int, maxValueLength int) (*Coffer, error, error) {
	return Db(dirPath).BatchSize(2000).
		LimitRecordsPerLogfile(5).
		FollowPause(400 * time.Millisecond).
		LogsByCheckpoint(2).
		MaxKeyLength(maxKeyLength).
		MaxValueLength(maxValueLength).
		MaxRecsPerOperation(10).
		Create()
}

func createNewCofferLength4T(maxKeyLength int, maxValueLength int) (*Coffer, error, error) {
	hdlExch := domain.Handler(handlerExchange)
	return Db(dirPath).BatchSize(2000).
		LimitRecordsPerLogfile(5).
		FollowPause(400 * time.Millisecond).
		LogsByCheckpoint(2).
		MaxKeyLength(maxKeyLength).
		MaxValueLength(maxValueLength).
		MaxRecsPerOperation(10).
		Handlers(map[string]*domain.Handler{"exchange": &hdlExch}).
		Create()
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
		if strings.HasSuffix(name, ".log") || strings.HasSuffix(name, ".check") || strings.HasSuffix(name, ".checkpoint") {
			os.Remove(dir + name)
		}
	}
	return nil
}

func handlerExchange(arg []byte, recs map[string][]byte) (map[string][]byte, error) {
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
