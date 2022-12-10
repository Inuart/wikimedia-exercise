package shortdescription

import (
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type cachedElement struct {
	insertion time.Time
	value     string
}

type cache struct {
	lru *lru.Cache[string, cachedElement]
	ttl time.Duration
}

func newCache(size int, ttl time.Duration) (cache, error) {
	c, err := lru.New[string, cachedElement](size)
	return cache{c, ttl}, err
}

func (c cache) Get(key string) (value string, ok bool) {
	elem, ok := c.lru.Get(key)
	if !ok || time.Since(elem.insertion) >= c.ttl {
		return "", false
	}

	return elem.value, true
}

func (c cache) Add(key, value string) {
	_ = c.lru.Add(key, cachedElement{
		insertion: time.Now(),
		value:     value,
	})
}
