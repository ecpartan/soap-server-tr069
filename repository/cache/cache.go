package repository

import "github.com/dgrijalva/lfu-go"

type Cache struct {
	c *lfu.Cache
}

func NewCache() *Cache {
	return &Cache{c: lfu.New()}
}
func (c *Cache) Get(key string) any {
	return c.c.Get(key)
}

func (c *Cache) Set(key string, value any) {
	c.c.Set(key, value)
}
