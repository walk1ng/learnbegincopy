package main

import (
	"fmt"
	"time"
	"wei/cpcache2go"
)

type myData struct {
	text     string
	moreData []byte
}

func main() {

	// will create the cache table for the first time accessing
	cache := cpcache2go.Cache("xyz")

	// put a new item in the cache and it will expire after not
	// being accessed for more than 5 seconds
	myval := myData{"This is a test!", []byte{}}
	cache.Add("myKey", 5*time.Second, &myval)

	// retrieve the item from the cache
	item, err := cache.Value("myKey")
	if err == nil {
		fmt.Println("find item in the cache and value:", item.Data().(*myData).text)
	} else {
		fmt.Println("error retrieve value from cache:", err)
	}

	// wait for the item to expire in cache
	time.Sleep(6 * time.Second)
	item, err = cache.Value("myKey")
	if err != nil {
		fmt.Println("item is not cached.")
	}

	// add another item that never expires
	cache.Add("myKey", 0, &myval)

	// add callback
	cache.SetAboutToDeleteItemCallback(func(item *cpcache2go.CacheItem) {
		fmt.Println("Deleting:", item.Key(), item.Data().(*myData).text, item.CreatedOn())
	})

	// remove the item from cache
	cache.Delete("myKey")

	// wipe the whole cache table
	cache.Flush()

}
