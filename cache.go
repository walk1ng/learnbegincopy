package cpcache2go

import (
	"sync"
)

var (
	cache = make(map[string]*CacheTable)
	mutex sync.RWMutex
)

// Cache return the existing cache table with the given name or creates a new one
// if the table does not exist
func Cache(table string) *CacheTable {
	mutex.RLock()
	t, ok := cache[table]
	mutex.RUnlock()

	if !ok {
		mutex.Lock()
		// double check if the table exists or not
		t, ok = cache[table]
		if !ok {
			t = &CacheTable{
				name:  table,
				items: make(map[interface{}]*CacheItem),
			}
			cache[table] = t
		}
		mutex.Unlock()
	}

	return t
}
