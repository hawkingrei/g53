package cache

import (
	"container/list"
	"github.com/hawkingrei/g53/servers"
	"math/rand"
	"reflect"
	"sync"
	"time"
)

type entry struct {
	Aliases    string
	RecordType string
	Value      string
	Private    bool
	TTL        int
	Time       time.Time
}
type Record struct {
	size int
	list []*list.Element
}
type RecordCache struct {
	mu   sync.RWMutex
	size uint64

	table map[string]*Record
}
type LRUCache struct {
	mu sync.RWMutex

	// list & table of *entry objects
	list  *list.List
	table map[string]*RecordCache

	// Our current size, in bytes. Obviously a gross simplification and low-grade
	// approximation.
	size uint64

	// How many bytes we are limiting the cache to.
	capacity uint64
}

//name to type
func NewLRUCache(capacity uint64) *LRUCache {
	return &LRUCache{
		list:     list.New(),
		table:    make(map[string]*RecordCache),
		capacity: capacity,
	}
}
func (lru *LRUCache) Get(s servers.Service) (e *entry, ok bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[s.Aliases]
	if element == nil {
		return &entry{}, false
	}
	record := element.table[s.RecordType]
	if record == nil {
		return &entry{}, false
	}
	return record.list[rand.Intn(len(record.list))].Value.(*entry), true
}
func (lru *LRUCache) Add(s servers.Service) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elements := lru.table[s.Aliases]; elements != nil {
		if element := elements.table[s.RecordType]; element == nil {
			Records := &Record{0, make([]*list.Element, 0)}
			elements.table[s.RecordType] = Records
		}
		content := &entry{s.Aliases, s.RecordType, s.Value, s.Private, s.TTL, time.Now()}
		elements.table[s.RecordType].list = append(elements.table[s.RecordType].list, lru.list.PushFront(content))
		lru.size = lru.size + 1
		if s.Private {
			lru.removeAllPublic(s)
		}
		lru.checkCapacity()
	} else {
		lru.addNew(s)
	}

}

func (lru *LRUCache) Set(originalValue servers.Service, modifyValue servers.Service) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if element := lru.table[originalValue.Aliases]; element != nil {
		tmp := element.table[originalValue.RecordType]
		for v := range tmp.list {

			if reflect.DeepEqual(originalValue.Value, tmp.list[v].Value.(*entry).Value) {
				tmp.list[v].Value.(*entry).TTL = modifyValue.TTL
				tmp.list[v].Value.(*entry).Private = modifyValue.Private
				tmp.list[v].Value.(*entry).Time = time.Now()
				tmp.list[v].Value.(*entry).Value = modifyValue.Value
				break
			}
		}
		if modifyValue.Private {
			lru.removeAllPublic(modifyValue)
		}
		return true
	}
	return false
}
func (lru *LRUCache) removeAllPublic(s servers.Service) bool {
	if element := lru.table[s.Aliases]; element != nil {
		tmp := element.table[s.RecordType].list
		for v := 0; v < len(tmp); v++ {
			if !tmp[v].Value.(*entry).Private {
				lru.list.Remove(tmp[v])
				tmp = append(tmp[:v], tmp[v+1:]...)
				v = v - 1
				lru.size = lru.size - 1
			}
		}
		return true
	}
	return false
}
func (lru *LRUCache) Remove(s servers.Service) bool {
	if element := lru.table[s.Aliases]; element != nil {
		tmp := element.table[s.RecordType].list
		for v := 0; v < len(tmp); v++ {
			if tmp[v].Value.(*entry).Value == s.Value {
				lru.list.Remove(tmp[v])
				tmp = append(tmp[:v], tmp[v+1:]...)
				v = v - 1
				lru.size = lru.size - 1

				if len(tmp) == 0 {
					delete(element.table, s.RecordType)
				}
				return true
				break
			}
		}
	}
	return false
}

func (lru *LRUCache) List(s servers.Service) (result map[int]entry) {
	lru.mu.RLock()
	defer lru.mu.RUnlock()
	element := lru.table[s.Aliases]
	if element == nil {
		return nil
	}
	record := (*element).table[s.RecordType]
	if record == nil {
		return nil
	}
	result = make(map[int]entry)
	for v := range record.list {
		result[v] = *(record.list[v].Value.(*entry))
	}
	return result
}
func (lru *LRUCache) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	lru.list.Init()
	lru.table = make(map[string]*RecordCache)
	lru.size = 0
}
func (lru *LRUCache) addNew(s servers.Service) {
	entries := &entry{s.Aliases, s.RecordType, s.Value, s.Private, s.TTL, time.Now()}
	Records := &Record{0, make([]*list.Element, 0)}
	(*Records).list = append((*Records).list, lru.list.PushFront(entries))
	RecordCaches := &RecordCache{size: 0, table: make(map[string]*Record)}
	RecordCaches.table[s.RecordType] = Records
	lru.table[s.Aliases] = RecordCaches
	lru.size = lru.size + 1
	lru.checkCapacity()
}
func (lru *LRUCache) checkCapacity() {
	for lru.size > lru.capacity {
		delElem := lru.list.Back()
		delValue := lru.table[delElem.Value.(*entry).Aliases]
		del := delValue.table[delElem.Value.(*entry).RecordType]

		for v := range del.list {
			if reflect.DeepEqual(delElem.Value.(*entry), del.list[v].Value.(*entry)) {
				del.list = append(del.list[:v], del.list[v+1:]...)
				//if len(del.list) ==0 {
				//	delete(delValue.table,delElem.Value.(*entry).RecordType)
				//}
				break
			}
		}

		lru.list.Remove(delElem)
		lru.size = lru.size - 1
	}
}
