package cache

import (
	"container/list"
    "sync"
)

type Service struct {
	//RecordType string
	Value      string
	TTL        int
	Aliases    string
}
type Record struct {
	size int
	list *list.List
}
type RecordCache struct {
	mu sync.Mutex

	list *list.List
	table map[string]*Record
}
type LRUCache struct {
	mu sync.Mutex

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

//type to record
func NewRecordCache() *RecordCache{
	return &RecordCache{
		list:	list.New(),
		table:  make(map[string]*Record,
	}
}