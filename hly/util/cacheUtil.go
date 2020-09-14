package util

import (
	"fmt"
	"github.com/muesli/cache2go"
	"time"
)

type myStruct struct {
	text     string
	moreData []byte
}

func CacheUtil() {
	cache := cache2go.Cache("myCache")
	val := myStruct{"This is a test!", []byte{}}
	cache.Add("someKey", 5*time.Second, &val)
	res, err := cache.Value("someKey")
	if err == nil {
		fmt.Println("Found value in cache:", res.Data().(*myStruct).text)
	} else {
		fmt.Println("Error retrieving value from cache:", err)
	}
}
