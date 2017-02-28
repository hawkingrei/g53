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

type EvictCallback func(s servers.Service)

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

//entry To server
func (c *LRU) entryToServer(s *entry) servers.Service {
	return servers.Service{(*s).RecordType, (*s).Value, (*s).TTL, false, (*s).Aliases}
}

// Purge is used to completely clear the cache
func (c *LRU) Purge() {
	for e := c.evictList.Front(); e != nil; e = e.Next() {
		if c.onEvict != nil {
			c.onEvict(c.entryToServer(e.Value.(*entry)))
		}
	}
	c.items = make(map[interface{}]*Records)
	c.evictList.Init()
}

func (c *LRU) Set(originalValue servers.Service, modifyValue servers.Service) error {
	if !(reflect.DeepEqual(originalValue.Aliases, modifyValue.Aliases) && reflect.DeepEqual(originalValue.RecordType, modifyValue.RecordType)) {
		return errors.New("Changed service's aliases and RecordType must be equal.")
	}
	if element := c.items[originalValue.Aliases]; element != nil {
		tmp := element.table[originalValue.RecordType]
		if tmp == nil {
			return errors.New("don't Exist service ")
		}
		for v := range tmp.list {
			if reflect.DeepEqual(originalValue.Value, tmp.list[v].Value.(*entry).Value) {
				tmp.list[v].Value.(*entry).TTL = modifyValue.TTL
				tmp.list[v].Value.(*entry).Time = time.Now()
				tmp.list[v].Value.(*entry).Value = modifyValue.Value
				return nil
			}
		}
	}
	return errors.New("don't Exist service ")
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
func (c *LRU) Get(s servers.Service) (servers.Service, error) {
	element := c.items[s.Aliases]
	if element == nil {
		return servers.Service{}, errors.New("Not exist")
	}
	record := element.table[s.RecordType]
	if record == nil {
		return servers.Service{}, errors.New("Not exist")
	}
	return c.entryToServer(record.list[rand.Intn(len(record.list))].Value.(*entry)), nil
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
	keys := make([]interface{}, c.evictList.Len())
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
func (c *LRU) Remove(s servers.Service) error {
	removeNum := 0
	if element := c.items[s.Aliases]; element != nil {
		tmp := element.table[s.RecordType].list
		for v := 0; v < len(tmp); v++ {
			if tmp[v].Value.(*entry).Value == s.Value {
				if c.onEvict != nil {
					c.onEvict(c.entryToServer(tmp[v].Value.(*entry)))
				}
				c.evictList.Remove(tmp[v])
				tmp = append(tmp[:v], tmp[v+1:]...)
				v = v - 1
				removeNum = removeNum + 1
				if len(tmp) == 0 {
					delete(element.table, s.RecordType)
				}

			}
		}
	}
	if removeNum > 0 {
		return nil
	}
	return errors.New("Nothing is removed")
}

// List list all element from the cache
func (c *LRU) List() []servers.Service {
	result := []servers.Service{}
	for aliases := range c.items {
		for recordType := range c.items[aliases].table {
			tmp := c.items[aliases].table[recordType].list
			for v := 0; v < len(tmp); v++ {
				result = append(result, tmp[v])
			}
		}
	}
	return result
}
