package cache

import (
	"errors"
	"github.com/hawkingrei/g53/cache/simplemsglru"
	"github.com/miekg/dns"
	"sync"
	"time"
)

func Round(val float64) uint32 {
    if val < 0 {
        return uint32(val-0.5)
    }
    return uint32(val+0.5)
}

// Cache is a thread-safe fixed size LRU cache.
type MsgCache struct {
	lru  *simplemsglru.LRU
	lock sync.RWMutex
}

// New creates an LRU of the given size
func NewMsgCache(size int) (*MsgCache, error) {
	return NewMsgCacheWithEvict(size, nil)
}

// NewWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewMsgCacheWithEvict(size int, onEvicted func(s *[]dns.RR)) (*MsgCache, error) {
	lru, err := simplemsglru.NewLRU(size, simplemsglru.EvictCallback(onEvicted))
	if err != nil {
		return nil, err
	}
	c := &MsgCache{
		lru: lru,
	}
	return c, nil
}

// Purge is used to completely clear the cache
func (c *MsgCache) Purge() {
	c.lock.Lock()
	c.lru.Purge()
	c.lock.Unlock()
}

// Get looks up a key's value from the cache.
func (c *MsgCache) Get(name string, rtype uint16) ([]dns.RR, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	result, rtime, err := c.lru.Get(name, rtype)

	if err != nil {
		return result, err
	}
	nowtime := time.Now()
	var rr []dns.RR = make([]dns.RR,len(result))
	for v := 0; v < len(result); v++ {
		expiration := rtime.Add(time.Duration(result[v].Header().Ttl) * time.Second)
		if expiration.Before(nowtime) {
			c.remove(name, rtype)
			return []dns.RR{}, errors.New("expiration")
		}
	}
	
	cttl := nowtime.Sub(*rtime).Seconds()
	for v := 0; v < len(result); v++ {
		rr[v] = dns.Copy(result[v])
	}


	for v := 0; v < len(rr); v++ {
		rr[v].Header().Ttl = rr[v].Header().Ttl - Round(cttl)
	}
	return rr, err
}



// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *MsgCache) Add(s []dns.RR) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Add(s)
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *MsgCache) Keys() []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.lru.Keys()
}

// Len returns the number of items in the cache.
func (c *MsgCache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.lru.Len()
}

func (c *MsgCache) remove(name string, rtype uint16) error {
	return c.lru.Remove(name, rtype)
}
