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
	Data []byte
}

type ReportReadList struct {
	Report
	Data     map[string][]byte
	NotFound []string
}

// type ReportDelete struct {
// 	Code  codes.Code
// 	Error error
// }

type ReportDeleteList struct {
	Report
	Removed  []string
	NotFound []string
}

type ReportRecordsCount struct {
	Report
	Count int
}
