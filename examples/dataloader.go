package main

import (
	"fmt"
	"strconv"
	"wei/cpcache2go"
)

func main() {
	cache := cpcache2go.Cache("myCache")

	// whenever someone tries to retrieve a non-existing key from cache
	// the data loader gets called automatically
	cache.SetDataLoader(func(key interface{}, args ...interface{}) *cpcache2go.CacheItem {
		// apply some loading logic here, e.g. read values for
		// this key from database, file or network
		value := "this is a mock value with key " + key.(string)
		// create a new item
		item := cpcache2go.NewCacheItem(key, 0, &value)

		return item
	})

	// add a item to the cache
	cache.Add("myKey", 0, "this is a test")

	// retrieve some item from the cache
	for i := 0; i < 10; i++ {
		res, err := cache.Value("someKey_" + strconv.Itoa(i))
		if err == nil {
			fmt.Println("found the value in cache:", res.Key())
		} else {
			fmt.Println("error retrieving value from cache:", err)
		}
	}

	cache.Foreach(func(key interface{}, item *cpcache2go.CacheItem) {
		fmt.Println("KEY:", key.(string))
	})

}
