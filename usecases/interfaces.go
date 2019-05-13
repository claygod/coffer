package usecases

// Coffer
// Interfaces
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

type Resourcer interface {
	GetPermission(int64) bool
}

type Porter interface {
	Catch([]string)
	Throw([]string)
}

type Logger interface {
	Write(error)
}
