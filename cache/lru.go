package cache

import (
	"github.com/hawkingrei/g53/cache/simplelru"
	"github.com/hawkingrei/g53/servers"
	"sync"
)

// Cache is a thread-safe fixed size LRU cache.
type Cache struct {
	lru  *simplelru.LRU
	lock sync.RWMutex
}

// New creates an LRU of the given size
func New(size int) (*Cache, error) {
	return NewWithEvict(size, nil)
}

// NewWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewWithEvict(size int, onEvicted func(s servers.Service)) (*Cache, error) {
	lru, err := simplelru.NewLRU(size, simplelru.EvictCallback(onEvicted))
	if err != nil {
		return nil, err
	}
	c := &Cache{
		lru: lru,
	}
	return c, nil
}

// Purge is used to completely clear the cache
func (c *Cache) Purge() {
	c.lock.Lock()
	c.lru.Purge()
	c.lock.Unlock()
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(s servers.Service) (servers.Service, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Get(s)
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *Cache) Add(s servers.Service) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Add(s)
}

// Set sets the provided key from the cache.
func (c *Cache) Set(originalValue servers.Service, modifyValue servers.Service) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Set(originalValue, modifyValue)
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(s servers.Service) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Remove(s)
	
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() {
	c.lock.Lock()
	c.lru.RemoveOldest()
	c.lock.Unlock()
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *Cache) Keys() []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.lru.Keys()
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.lru.Len()
}
