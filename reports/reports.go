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

/*
ReportRead - report returned after operation Read.
*/
type ReportRead struct {
	Report
	Data []byte
}

/*
ReportReadList - report returned after operation ReadList.
*/
type ReportReadList struct {
	Report
	Data     map[string][]byte
	NotFound []string
}

/*
ReportTransaction - report returned after operation Transaction.
*/
type ReportTransaction struct {
	Report
	Data map[string][]byte
}

/*
ReportRecordsList - report returned after operation RecordsList.
*/
type ReportRecordsList struct {
	Report
	Data []string
}

// type ReportDelete struct {
// 	Code  codes.Code
// 	Error error
// }

/*
ReportDeleteList - report returned after operation DeleteList.
*/
type ReportDeleteList struct {
	Report
	Removed  []string
	NotFound []string
}

/*
ReportRecordsCount - report returned after operation Count.
*/
type ReportRecordsCount struct {
	Report
	Count int
}

/*
ReportWriteList - report returned after operation WriteList Optional/Strict.
*/
type ReportWriteList struct {
	Report
	Found []string
}

/*
Report - report returned after operation.
*/
type Report struct {
	Code  codes.Code
	Error error
}

/*
IsCodeOk - checking code for Ok result
*/
func (r *Report) IsCodeOk() bool {
	return r.Code == codes.Ok
}

// func (r *Report) IsCodeWarning() bool {
// 	return r.Code >= codes.Warning
// }

/*
IsCodeError - checking code for availability `Error`
*/
func (r *Report) IsCodeError() bool {
	return r.Code > codes.Ok && r.Code < codes.Panic
}

/*
IsCodeErrRecordLimitExceeded - checking code for error ErrRecordLimitExceeded
*/
func (r *Report) IsCodeErrRecordLimitExceeded() bool {
	return r.Code == codes.ErrRecordLimitExceeded
}

/*
IsCodeErrExceedingMaxValueSize - checking code for error ErrExceedingMaxValueSize
*/
func (r *Report) IsCodeErrExceedingMaxValueSize() bool {
	return r.Code == codes.ErrExceedingMaxValueSize
}

/*
IsCodeErrExceedingMaxKeyLength - checking code for error ErrExceedingMaxKeyLength
*/
func (r *Report) IsCodeErrExceedingMaxKeyLength() bool {
	return r.Code == codes.ErrExceedingMaxKeyLength
}

/*
IsCodeErrExceedingZeroKeyLength - checking code for error ErrExceedingZeroKeyLength
*/
func (r *Report) IsCodeErrExceedingZeroKeyLength() bool {
	return r.Code == codes.ErrExceedingZeroKeyLength
}

/*
IsCodeErrHandlerNotFound - checking code for error ErrHandlerNotFound
*/
func (r *Report) IsCodeErrHandlerNotFound() bool {
	return r.Code == codes.ErrHandlerNotFound
}

/*
IsCodeErrParseRequest - checking code for error ErrParseRequest
*/
func (r *Report) IsCodeErrParseRequest() bool {
	return r.Code == codes.ErrParseRequest
}

/*
IsCodeErrResources - checking code for error ErrResources
*/
func (r *Report) IsCodeErrResources() bool {
	return r.Code == codes.ErrResources
}

/*
IsCodeErrNotFound - checking code for error ErrNotFound
*/
func (r *Report) IsCodeErrNotFound() bool {
	return r.Code == codes.ErrNotFound
}

/*
IsCodeErrRecordsFound - checking code for error ErrRecordsFound
*/
func (r *Report) IsCodeErrRecordsFound() bool {
	return r.Code == codes.ErrRecordsFound
}

/*
IsCodeErrReadRecords - checking code for error ErrReadRecords
*/
func (r *Report) IsCodeErrReadRecords() bool {
	return r.Code == codes.ErrReadRecords
}

/*
IsCodeErrHandlerReturn - checking code for error ErrHandlerReturn
*/
func (r *Report) IsCodeErrHandlerReturn() bool {
	return r.Code == codes.ErrHandlerReturn
}

/*
IsCodeErrHandlerResponse - checking code for error ErrHandlerResponse
*/
func (r *Report) IsCodeErrHandlerResponse() bool {
	return r.Code == codes.ErrHandlerResponse
}

/*
IsCodePanic - checking code for availability `Panic`
*/
func (r *Report) IsCodePanic() bool {
	return r.Code >= codes.Panic
}

/*
IsCodePanicStopped - checking code for error PanicStopped
*/
func (r *Report) IsCodePanicStopped() bool {
	return r.Code == codes.PanicStopped
}

/*
IsCodePanicWAL - checking code for error PanicWAL
*/
func (r *Report) IsCodePanicWAL() bool {
	return r.Code == codes.PanicWAL
}
