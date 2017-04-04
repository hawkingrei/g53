package cache

import (
	"github.com/hawkingrei/g53/cache/simplemsglru"
	"github.com/miekg/dns"
	"github.com/spaolacci/murmur3"
	"sync"
	"time"
)

func hashFunc(data []byte) uint64 {
	return murmur3.Sum64(data)
}

func Round(val float64) uint32 {
	if val < 0 {
		return uint32(val - 0.5)
	}
	return uint32(val + 0.5)
}

// Cache is a thread-safe fixed size LRU cache.
type MsgCache struct {
	lru  [256]*simplemsglru.LRU
	lock [256]sync.RWMutex
}

// New creates an LRU of the given size
func NewMsgCache(size int) (*MsgCache, error) {
	return NewMsgCacheWithEvict(size, nil)
}

// NewWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewMsgCacheWithEvict(size int, onEvicted func(s *[]dns.RR)) (c *MsgCache, err error) {
	c = new(MsgCache)
	for i := 0; i < 256; i++ {
		c.lru[i], err = simplemsglru.NewLRU(size/256, simplemsglru.EvictCallback(onEvicted))
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

// Purge is used to completely clear the cache
func (c *MsgCache) Purge() {
	for i := 0; i < 256; i++ {
		c.lock[i].Lock()
		c.lru[i].Purge()
		c.lock[i].Unlock()
	}
}

// Get looks up a key's value from the cache.
func (c *MsgCache) Get(name string, rtype uint16) ([]dns.RR, *time.Time, error) {
	hashVal := hashFunc([]byte(name))
	segId := hashVal & 255
	c.lock[segId].Lock()
	defer c.lock[segId].Unlock()
	return c.lru[segId].Get(name, rtype)
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *MsgCache) Add(s []dns.RR, rtype uint16) bool {
	hashVal := hashFunc([]byte(s[0].Header().Name))
	segId := hashVal & 255
	c.lock[segId].Lock()
	defer c.lock[segId].Unlock()
	return c.lru[segId].Add(s, rtype)
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *MsgCache) Keys() (result []interface{}) {
	for i := 0; i < 256; i++ {
		c.lock[i].RLock()
		result = append(result, c.lru[i].Keys()...)
		c.lock[i].RUnlock()
	}
	return result
}

// Len returns the number of items in the cache.
func (c *MsgCache) Len() (result int) {
	for i := 0; i < 256; i++ {
		c.lock[i].RLock()
		result = result + c.lru[i].Len()
		c.lock[i].RUnlock()
	}
	return result
}

func (c *MsgCache) Remove(name string, rtype uint16) error {
	hashVal := hashFunc([]byte(name))
	segId := hashVal & 255
	c.lock[segId].Lock()
	defer c.lock[segId].Unlock()
	return c.lru[segId].Remove(name, rtype)
}
