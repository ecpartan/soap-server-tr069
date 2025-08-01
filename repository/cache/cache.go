package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ecpartan/soap-server-tr069/internal/config"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/go-redis/redis/v8"
	"github.com/m7shapan/lfu-redis"
)

type Cache struct {
	rediscli *redis.Client
	c        *lfu.LFUCache
	sync.RWMutex
}

var c *Cache

func NewCache(ctx context.Context, cfg *config.Config) *Cache {

	once := &sync.Once{}
	once.Do(func() {
		redisClient := redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password:     cfg.Redis.Password,
			DB:           cfg.Redis.DB,
			PoolSize:     cfg.Redis.PoolSize,
			MinIdleConns: cfg.Redis.MinIdleConns,
		})

		c = &Cache{c: lfu.New(cfg.Redis.MaxActiveConns, redisClient), rediscli: redisClient}
	})
	return c
}

func GetCache() *Cache {
	return c
}

func (c *Cache) Get(key string) map[string]any {
	c.RLock()
	defer c.RUnlock()
	if val, err := c.c.Get(key); err == nil {
		ret := make(map[string]any)
		json.Unmarshal([]byte(val), &ret)
		return ret
	} else {
		logger.LogDebug("Error getting value from cache: %v", err)
		return nil
	}
}

func (c *Cache) Set(key string, value any) {
	ret, _ := json.Marshal(value)
	logger.LogDebug("Setting cache value: %s", string(ret))
	c.Lock()
	defer c.Unlock()
	err := c.c.Put(key, ret)
	if err != nil {
		logger.LogDebug("Error getting value from cache: %v", err)
	}
}
