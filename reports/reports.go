package reports

// Coffer
// Reports (level usecases)
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"

	"github.com/claygod/coffer/reports/codes"
)

type ReportRead struct {
	Code  codes.Code
	Data  []byte
	Error error
}

type ReportReadList struct {
	Code     codes.Code
	Data     map[string][]byte
	NotFound []string
	Error    error
}
