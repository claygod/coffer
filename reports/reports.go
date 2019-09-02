package reports

// Coffer
// Reports (level usecases)
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"

	"github.com/claygod/coffer/reports/codes"
)

type Report struct {
	Code  codes.Code
	Error error
}

// type ReportTransaction struct {
// 	Code  codes.Code
// 	Error error
// }

// type ReportWriteList struct {
// 	Code  codes.Code
// 	Error error
// }

type ReportRead struct {
	Report
	//Code  codes.Code
	Data []byte
	//Error error
}

type ReportReadList struct {
	Report
	//Code     codes.Code
	Data     map[string][]byte
	NotFound []string
	//Error    error
}

// type ReportDelete struct {
// 	Code  codes.Code
// 	Error error
// }

type ReportDeleteList struct {
	Report
	//Code     codes.Code
	Removed  []string
	NotFound []string
	//Error    error
}
