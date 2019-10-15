package reports

// Coffer
// Reports (level usecases)
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	//"fmt"

	"github.com/claygod/coffer/reports/codes"
)

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

type ReportTransaction struct {
	Report
	Data map[string][]byte
}

type ReportRecordsList struct {
	Report
	Data []string
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

type Report struct {
	Code  codes.Code
	Error error
}

func (r *Report) IsCodeOk() bool {
	return r.Code == codes.Ok
}

// func (r *Report) IsCodeWarning() bool {
// 	return r.Code >= codes.Warning
// }
func (r *Report) IsCodeError() bool {
	return r.Code > codes.Ok && r.Code < codes.Panic
}
func (r *Report) IsCodeErrRecordLimitExceeded() bool {
	return r.Code == codes.ErrRecordLimitExceeded
}
func (r *Report) IsCodeErrExceedingMaxValueSize() bool {
	return r.Code == codes.ErrExceedingMaxValueSize
}
func (r *Report) IsCodeErrExceedingMaxKeyLength() bool {
	return r.Code == codes.ErrExceedingMaxKeyLength
}
func (r *Report) IsCodeErrExceedingZeroKeyLength() bool {
	return r.Code == codes.ErrExceedingZeroKeyLength
}
func (r *Report) IsCodeErrHandlerNotFound() bool {
	return r.Code == codes.ErrHandlerNotFound
}
func (r *Report) IsCodeErrParseRequest() bool {
	return r.Code == codes.ErrParseRequest
}
func (r *Report) IsCodeErrResources() bool {
	return r.Code == codes.ErrResources
}
func (r *Report) IsCodeErrNotFound() bool {
	return r.Code == codes.ErrNotFound
}
func (r *Report) IsCodeErrReadRecords() bool {
	return r.Code == codes.ErrReadRecords
}
func (r *Report) IsCodeErrHandlerReturn() bool {
	return r.Code == codes.ErrHandlerReturn
}
func (r *Report) IsCodeErrHandlerResponse() bool {
	return r.Code == codes.ErrHandlerResponse
}
func (r *Report) IsCodePanic() bool {
	return r.Code >= codes.Panic
}
func (r *Report) IsCodePanicStopped() bool {
	return r.Code == codes.PanicStopped
}
func (r *Report) IsCodePanicWAL() bool {
	return r.Code == codes.PanicWAL
}
