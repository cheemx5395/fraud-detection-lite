package helpers

import "sync"

var (
	TokenList      = make(map[string]struct{})
	WhiteListMutex = &sync.Mutex{}
)
