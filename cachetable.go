package cpcache2go

import (
	"log"
	"sync"
	"time"
)

// CacheTable is a table in the cache
type CacheTable struct {
	sync.RWMutex

	// the table's name
	name string
	// all cached items
	items map[interface{}]*CacheItem

	// timer responsible for triggering cleanup
	cleanupTimer time.Timer
	// current timer duration
	cleanupInterval time.Duration

	// logger for the talbe
	logger *log.Logger

	// callback method triggered when trying to load a non-existing key
	loadData func(key interface{}, args ...interface{}) *CacheItem
	// callback method triggered when adding a new item to the cache
	addedItem func(item *CacheItem)
	// callback method triggered before deleting an item from the cache
	aboutToDeleteItem func(item *CacheItem)
}

// Count return how many items are currently stored in the cache
func (table *CacheTable) Count() int {
	table.RLock()
	defer table.RUnlock()
	return len(table.items)
}

// Foreach all items in the table
func (table *CacheTable) Foreach(trans func(k interface{}, item *CacheItem)) {
	table.RLock()
	defer table.RUnlock()

	for k, v := range table.items {
		trans(k, v)
	}
}

// SetDataLoader configure a data-loader callback, which will be called when
// trying to access a non-exisiting key. The key and 0...n additional arguments
// are passed to the callback function
func (table *CacheTable) SetDataLoader(f func(interface{}, ...interface{}) *CacheItem) {
	table.Lock()
	defer table.Unlock()
	table.loadData = f
}

// SetAddedItemCallback configure a callback, which will be called when
// a new item is added to the cache
func (table *CacheTable) SetAddedItemCallback(f func(item *CacheItem)) {
	table.Lock()
	defer table.Unlock()
	table.addedItem = f
}

// SetLogger configure the logger used by the table
func (table *CacheTable) SetLogger(logger *log.Logger) {
	table.Lock()
	defer table.Unlock()
	table.logger = logger
}
