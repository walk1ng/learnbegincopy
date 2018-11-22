package cpcache2go

import (
	"log"
	"sort"
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
	cleanupTimer *time.Timer
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

// SetAboutToDeleteItemCallback configures a callback, which will be called
// every time an item is about to removed from the cache
func (table *CacheTable) SetAboutToDeleteItemCallback(f func(item *CacheItem)) {
	table.Lock()
	defer table.Unlock()
	table.aboutToDeleteItem = f
}

// SetLogger configure the logger used by the table
func (table *CacheTable) SetLogger(logger *log.Logger) {
	table.Lock()
	defer table.Unlock()
	table.logger = logger
}

// expiration check loop, triggered by a self-adjusting timer
func (table *CacheTable) expirationCheck() {
	table.Lock()
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	if table.cleanupInterval > 0 {
		table.log("Expiration check trigged after", table.cleanupInterval, "for table", table.name)
	} else {
		table.log("Expiration check installed for table", table.name)
	}

	now := time.Now()
	smallestDuration := 0 * time.Second
	for key, item := range table.items {
		item.RLock()
		lifeSpan := item.lifeSpan
		accessedOn := item.accessedOn
		item.RUnlock()

		if lifeSpan == 0 {
			continue
		}
		if now.Sub(accessedOn) >= lifeSpan {
			// item has exceeded its lifespan
			table.deleteInternal(key)
		} else {
			// find the item chronologically closest to its end-of-lifespan
			if smallestDuration == 0 || lifeSpan-now.Sub(accessedOn) < smallestDuration {
				smallestDuration = lifeSpan - now.Sub(accessedOn)
			}
		}
	}

	// setup the interval for the next cleanup check
	table.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go table.expirationCheck()
		})
	}
	table.Unlock()
}

// add item to the cache, the method is internal
func (table *CacheTable) addInternal(item *CacheItem) {
	table.log("Adding item with key", item.key, "and lifespan of", item.lifeSpan, "to table", table.name)
	table.items[item.key] = item

	// cache value so we don't keep blocking the mutex
	expDur := table.cleanupInterval
	addedItem := table.addedItem
	table.Unlock()

	// Trigger callback after adding the item to cache
	if addedItem != nil {
		addedItem(item)
	}

	// If we haven't set up any expiration check timer or found a more imminent item
	if item.lifeSpan > 0 && (expDur == 0 || item.lifeSpan < expDur) {
		table.expirationCheck()
	}
}

// Add adds a key/value pair to the cache
func (table *CacheTable) Add(key interface{}, lifeSpan time.Duration, data interface{}) *CacheItem {
	item := NewCacheItem(key, lifeSpan, data)
	// Add item to the cache
	table.Lock()
	table.addInternal(item)

	return item
}

// delete item from the cache, the method is internal
func (table *CacheTable) deleteInternal(key interface{}) (*CacheItem, error) {
	r, ok := table.items[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	// cache value so we don't keep blocking the mutex
	aboutToDeleteItem := table.aboutToDeleteItem
	table.Unlock()

	// trigger the callback before deleting the item from cache
	if aboutToDeleteItem != nil {
		aboutToDeleteItem(r)
	}

	r.RLock()
	defer r.RUnlock()
	if r.aboutToExpire != nil {
		r.aboutToExpire(key)
	}

	table.Lock()
	table.log("Deleting item with key", key, "was created on", r.CreatedOn(), "and hit", r.AccessCount(), "times from table", table.name)
	delete(table.items, key)

	return r, nil
}

// Delete item from the cache, the method is exported
func (table *CacheTable) Delete(key interface{}) (*CacheItem, error) {
	table.Lock()
	defer table.Unlock()

	return table.deleteInternal(key)
}

// Exists returns if an item exists in the cache but doesn't
// try to fetch data via the loadData callback
func (table *CacheTable) Exists(key interface{}) bool {
	table.RLock()
	defer table.RUnlock()
	_, ok := table.items[key]

	return ok
}

// NotFoundAdd tests whether an item not found in the cache. Unlike the Exists
// method this also adds data if the key could not be found.
func (table *CacheTable) NotFoundAdd(key interface{}, lifeSpan time.Duration, data interface{}) bool {
	table.Lock()

	if _, ok := table.items[key]; ok {
		table.Unlock()
		return false
	}

	item := NewCacheItem(key, lifeSpan, data)
	table.addInternal(item)

	return true
}

// Value returns an item from the cache and marks it to be kept alive.
// You can pass additional arguments to your Dataloader callback function.
func (table *CacheTable) Value(key interface{}, args ...interface{}) (*CacheItem, error) {
	table.RLock()
	r, ok := table.items[key]
	loadData := table.loadData
	table.RUnlock()

	if ok {
		// update access counter and timestamp
		r.KeepAlive()
		return r, nil
	}

	// item doesn't exist in the cache. Try and fetch it with a data-loader
	if loadData != nil {
		item := loadData(key, args...)
		if item != nil {
			table.Add(key, item.lifeSpan, item.data)
			return item, nil
		}
		return nil, ErrKeyNotFoundOrLoadable
	}

	return nil, ErrKeyNotFound
}

// Flush deletes all items in cache
func (table *CacheTable) Flush() {
	table.Lock()
	defer table.Unlock()

	table.log("Flushing table", table.name)
	table.items = make(map[interface{}]*CacheItem)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
}

// CacheItemPair maps key to access counter
type CacheItemPair struct {
	Key         interface{}
	AccessCount int64
}

// CacheItemPairList is a slice of CacheItemPairs that implements sort
// Sort by AccessCount
type CacheItemPairList []CacheItemPair

// Swap method for CacheItemPairList
func (p CacheItemPairList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Len method for CacheItemPairList
func (p CacheItemPairList) Len() int {
	return len(p)
}

// Less method for CacheItemPairList
func (p CacheItemPairList) Less(i, j int) bool {
	return p[i].AccessCount > p[j].AccessCount
}

// MostAccessed returns the most accessed item in this cache table
func (table *CacheTable) MostAccessed(count int64) []*CacheItem {
	table.RLock()
	defer table.RUnlock()

	p := make(CacheItemPairList, len(table.items))
	i := 0
	for k, v := range table.items {
		p[i] = CacheItemPair{
			k,
			v.accessCount,
		}
		i++
	}
	sort.Sort(p)

	var r []*CacheItem
	c := int64(0)
	for _, v := range p {
		if c >= count {
			break
		}
		item, ok := table.items[v.Key]
		if ok {
			r = append(r, item)
		}
		c++
	}

	return r
}

// Internal logging method for convenience
func (table *CacheTable) log(v ...interface{}) {
	if table.logger == nil {
		return
	}
	table.logger.Println(v...)
}
