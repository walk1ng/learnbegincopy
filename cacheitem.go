package cpcache2go

import (
	"sync"
	"time"
)

// CacheItem is an individual cache item
// Parameter data contains the user-set value in the cache
type CacheItem struct {
	sync.RWMutex

	// item's key
	key interface{}
	// item's data
	data interface{}
	// how long will the item live in the cache when not being acceseed/kept alive
	lifeSpan time.Duration

	// creation timestamp
	createdOn time.Time
	// last access timestamp
	accessedOn time.Time
	// how often the item was accessed
	accessCount int64

	// callback method triggered right before removing the item from the cache
	aboutToExpire func(key interface{})
}

// NewCacheItem return a newly created CacheItem
func NewCacheItem(key interface{}, lifeSpan time.Duration, data interface{}) *CacheItem {
	t := time.Now()
	return &CacheItem{
		key:           key,
		data:          data,
		lifeSpan:      lifeSpan,
		createdOn:     t,
		accessedOn:    t,
		accessCount:   0,
		aboutToExpire: nil,
	}
}

// KeepAlive marks an item to be kept for another expireDuration period
func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedOn = time.Now()
	item.accessCount++
}

// LifeSpan returns the item's expiration duration
func (item *CacheItem) LifeSpan() time.Duration {
	// immutable
	return item.lifeSpan
}

// AccessedOn return when the item was last accessed
func (item *CacheItem) AccessedOn() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessedOn
}

// CreatedOn return when the item was added to the cache
func (item *CacheItem) CreatedOn() time.Time {
	// immutable
	return item.createdOn
}

// AccessCount return how often the item has been accessed
func (item *CacheItem) AccessCount() int64 {
	item.RLock()
	defer item.RUnlock()
	return item.accessCount
}

// Key return the key of this cached item
func (item *CacheItem) Key() interface{} {
	// immutable
	return item.key
}

// Data return the data of this cached item
func (item *CacheItem) Data() interface{} {
	// immutable
	return item.data
}

// SetAboutToExpireCallback configure a callback, which will be called right
// before the item is about to be removed from the cache
func (item *CacheItem) SetAboutToExpireCallback(f func(interface{})) {
	item.Lock()
	defer item.Unlock()
	item.aboutToExpire = f
}
