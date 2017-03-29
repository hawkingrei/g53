package simplemsglru

import (
	"container/list"
	"errors"
	"github.com/miekg/dns"
	"time"
)

type EvictCallback func(s *[]dns.RR)

type Record struct {
	list *list.Element
	Time time.Time
}

type Records struct {
	table map[interface{}]*Record
}

type LRU struct {
	size      int
	evictList *list.List
	items     map[interface{}]*Records
	onEvict   EvictCallback
}

func NewLRU(size int, onEvict EvictCallback) (*LRU, error) {
	if size <= 0 {
		return nil, errors.New("Must provide a positive size")
	}
	c := &LRU{
		size:      size,
		evictList: list.New(),
		items:     make(map[interface{}]*Records),
		onEvict:   onEvict,
	}
	return c, nil
}

// Purge is used to completely clear the cache
func (c *LRU) Purge() {
	for e := c.evictList.Front(); e != nil; e = e.Next() {
		if c.onEvict != nil {
			c.onEvict(e.Value.(*[]dns.RR))
		}
	}
	c.items = make(map[interface{}]*Records)
	c.evictList.Init()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *LRU) Add(s []dns.RR) bool {
	if len(s) == 0 {
		return false
	}
	rrtype := s[0].Header().Rrtype
	name := s[0].Header().Name
	if elements := c.items[name]; elements != nil {
		if element := elements.table[rrtype]; element == nil {
			if len(elements.table) == 0 {
				Records := &Record{}
				elements.table[rrtype] = Records
			} else {
				for rt := range elements.table {
					if (rt == dns.TypeA || rt == dns.TypeAAAA) && (rrtype == dns.TypeCNAME) {
						return false
					}
					if (rrtype == dns.TypeA || rrtype == dns.TypeAAAA) && (rt == dns.TypeCNAME) {
						return false
					} else {
						Records := &Record{}
						elements.table[rrtype] = Records
					}
				}
			}
		}
		elements.table[rrtype].list = c.evictList.PushFront(&s)
		elements.table[rrtype].Time = time.Now()
	} else {
		c.addNew(s)
	}
	evict := c.evictList.Len() > c.size
	if evict {
		c.RemoveOldest()
	}
	return evict
}

func (c *LRU) addNew(s []dns.RR) {
	rrtype := s[0].Header().Rrtype
	name := s[0].Header().Name
	entries := &s
	newRecord := &Record{}
	(*newRecord).list = c.evictList.PushFront(entries)
	(*newRecord).Time = time.Now()
	newRecords := &Records{table: make(map[interface{}]*Record)}
	newRecords.table[rrtype] = newRecord
	c.items[name] = newRecords
}

// Get looks up a key's value from the cache.
func (c *LRU) Get(name string, rtype uint16) ([]dns.RR, *time.Time, error) {
	element := c.items[name]
	if element == nil {
		return []dns.RR{}, &time.Time{}, errors.New("Not exist")
	}
	record := element.table[rtype]
	if record == nil {
		return []dns.RR{}, &time.Time{}, errors.New("Not exist")
	}
	return *(record.list.Value.(*[]dns.RR)), &record.Time, nil
}

// Check if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRU) Contains(key interface{}) (ok bool) {
	_, ok = c.items[key]
	return ok
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRU) RemoveOldest() {
	delElem := c.evictList.Back()
	if delElem == nil {
		return
	}
	del := delElem.Value.(*[]dns.RR)
	name := (*del)[0].Header().Name
	rtype := (*del)[0].Header().Rrtype
	delValue := c.items[name]
	delete(delValue.table, rtype)
	c.evictList.Remove(delElem)
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRU) Keys() []interface{} {
	keys := make([]interface{}, c.evictList.Len())
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		content := ent.Value.(*[]dns.RR)
		keys[i] = (*content)[0].Header().Name
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
	return c.evictList.Len()
}

func (c *LRU) Remove(name string, rtype uint16) error {
	element := c.items[name]
	if element == nil {
		return errors.New("Not exist")
	}
	record := element.table[rtype]
	if record == nil {
		return errors.New("Not exist")
	}
	c.evictList.Remove(record.list)
	delete(element.table, rtype)
	if len(element.table) == 0 {
		delete(c.items, name)
	}
	return nil
}
