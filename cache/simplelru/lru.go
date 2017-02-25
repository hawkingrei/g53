package simplelru

import (
	"container/list"
	"errors"
	"github.com/hawkingrei/g53/servers"
	"math/rand"
	"reflect"
	"time"
)

type entry struct {
	Aliases    string
	RecordType string
	Value      string
	TTL        int
	Time       time.Time
}

type EvictCallback func(s *entry)

type Record struct {
	list []*list.Element
}

type Records struct {
	table map[string]*Record
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
			c.onEvict(e.Value.(*entry))
		}
	}
	c.items = make(map[interface{}]*Records)
	c.evictList.Init()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *LRU) Add(s servers.Service) bool {
	if elements := c.items[s.Aliases]; elements != nil {
		if element := elements.table[s.RecordType]; element == nil {
			Records := &Record{make([]*list.Element, 0)}
			elements.table[s.RecordType] = Records
		}
		content := &entry{s.Aliases, s.RecordType, s.Value, s.TTL, time.Now()}
		elements.table[s.RecordType].list = append(elements.table[s.RecordType].list, c.evictList.PushFront(content))
	} else {
		c.addNew(s)
	}
	evict := c.evictList.Len() > c.size
	if evict {
		c.RemoveOldest()
	}
	return evict
}

// Get looks up a key's value from the cache.
func (c *LRU) Get(s servers.Service) (*entry, error) {
	element := c.items[s.Aliases]
	if element == nil {
		return &entry{}, errors.New("Not exist")
	}
	record := element.table[s.RecordType]
	if record == nil {
		return &entry{}, errors.New("Not exist")
	}
	return record.list[rand.Intn(len(record.list))].Value.(*entry), nil
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
	delValue := c.items[delElem.Value.(*entry).Aliases]
	del := delValue.table[delElem.Value.(*entry).RecordType]

	for v := range del.list {
		if reflect.DeepEqual(delElem.Value.(*entry), del.list[v].Value.(*entry)) {
			del.list = append(del.list[:v], del.list[v+1:]...)
			if len(del.list) == 0 {
				delete(delValue.table, delElem.Value.(*entry).RecordType)
			}
			break
		}
	}
	c.evictList.Remove(delElem)
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRU) Keys() []interface{} {
	keys := make([]interface{}, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry).Aliases
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
	return c.evictList.Len()
}

func (c *LRU) addNew(s servers.Service) {
	entries := &entry{s.Aliases, s.RecordType, s.Value, s.TTL, time.Now()}
	newRecord := &Record{make([]*list.Element, 0)}
	(*newRecord).list = append((*newRecord).list, c.evictList.PushFront(entries))
	newRecords := &Records{table: make(map[string]*Record)}
	newRecords.table[s.RecordType] = newRecord
	c.items[s.Aliases] = newRecords
}

// removeElement is used to remove a given list element from the cache
func (c *LRU) Remove(s servers.Service) {
	if element := c.items[s.Aliases]; element != nil {
		tmp := element.table[s.RecordType].list
		for v := 0; v < len(tmp); v++ {
			if tmp[v].Value.(*entry).Value == s.Value {
				c.evictList.Remove(tmp[v])
				tmp = append(tmp[:v], tmp[v+1:]...)
				v = v - 1
				if len(tmp) == 0 {
					delete(element.table, s.RecordType)
				}
				if c.onEvict != nil {
					c.onEvict(&entry{s.Aliases, s.RecordType, s.Value, s.TTL, time.Now()})
				}
				break
			}
		}
	}
}
