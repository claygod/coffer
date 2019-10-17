package main

// Coffer
// Example quik start handler
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
)

/*
HandlerExchange - exchange handler.
*/
func HandlerExchange(arg []byte, recs map[string][]byte) (map[string][]byte, error) {
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
