package main

import (
	"fmt"
	"time"
	"wei/cpcache2go"
)

func main() {
	cache := cpcache2go.Cache("myCache")

	// add a callback and which will be triggered every time
	// a new item is added to the cache
	cache.SetAddedItemCallback(func(item *cpcache2go.CacheItem) {
		fmt.Println("Added:", item.Key(), item.Data(), item.CreatedOn())
	})

	// add a callback and which will be triggered every time
	// a new item is about to remove from the cache
	cache.SetAboutToDeleteItemCallback(func(item *cpcache2go.CacheItem) {
		fmt.Println("Deleting:", item.Key(), item.Data(), item.CreatedOn(), item.AccessedOn())
	})

	// caching a new item will trigger the AddedItem callback
	cache.Add("someKey", 0, "this is data")

	// wait 3 seconds then to retrieve the item from the cache
	time.Sleep(3 * time.Second)
	res, err := cache.Value("someKey")
	if err == nil {
		fmt.Println("found the value from cache:", res.Data())
	} else {
		fmt.Println("error retrieving value from cache:", err)
	}

	// bug?
	// if registry the expire callback for a item which never be expired
	res.SetAboutToExpireCallback(func(key interface{}) {
		fmt.Println("Are you sure i am expire?:", key.(string))
	})

	// deleting the item will trigger the AboutToDeleteItem callback
	cache.Delete("someKey")

	// caching a new item in cache
	res = cache.Add("anotherKey", 3*time.Second, "this is another test")

	// add a callback and which will be triggeredd every time
	// the item is about to expire
	res.SetAboutToExpireCallback(func(key interface{}) {
		fmt.Println("About to expire:", key.(string))
	})

	time.Sleep(5 * time.Second)

}
