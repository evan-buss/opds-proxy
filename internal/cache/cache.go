package cache

import (
	"sync"
	"time"
)

type CacheEntry[T any] struct {
	timestamp time.Time
	Value     *T
}

type Cache[T any] struct {
	entries map[string]*CacheEntry[T]
	config  CacheConfig
	mutex   sync.Mutex
}

type CacheConfig struct {
	TTL             time.Duration
	CleanupInterval time.Duration
}

func NewCache[T any](config CacheConfig) *Cache[T] {
	cache := &Cache[T]{
		entries: make(map[string]*CacheEntry[T]),
		config:  config,
	}
	go cache.cleanupLoop()
	return cache
}

func (c *Cache[T]) Set(key string, entry *T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries[key] = &CacheEntry[T]{timestamp: time.Now(), Value: entry}
}

func (c *Cache[T]) Get(key string) (*T, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if exists && time.Since(entry.timestamp) > c.config.TTL {
		delete(c.entries, key)
		return nil, false
	}
	return entry.Value, exists
}

func (c *Cache[T]) cleanupLoop() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanEntries()
	}
}

func (c *Cache[T]) cleanEntries() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for key, entry := range c.entries {
		if time.Since(entry.timestamp) > c.config.TTL {
			delete(c.entries, key)
		}
	}
}
