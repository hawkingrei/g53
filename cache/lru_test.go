package cache

import (
	"fmt"
	"github.com/hawkingrei/g53/servers"
	"testing"
)

func TestCache(t *testing.T) {
	l := NewLRUCache(3)
	l.Add(servers.Service{"A", "10.0.0.0", 600, false, "www.google.com"})
	l.Get(servers.Service{"A", "", 0, false, "www.google.com"})
	if tmp, _ := l.Get(servers.Service{"MX", "", 0, false, "www.google.com"}); (*tmp != entry{}) {
		t.Errorf("not get nil")
	}
	if tmp, _ := l.Get(servers.Service{"MX", "", 0, false, "www.taobao.com"}); (*tmp != entry{}) {
		t.Errorf("not get nil")
	}
	fmt.Println(l.List(servers.Service{"A", "", 0, false, "www.google.com"}))
	l.Set(servers.Service{"A", "10.0.0.0", 500, false, "www.google.com"}, servers.Service{"A", "10.0.0.1", 500, false, "www.google.com"})
	l.Add(servers.Service{"A", "10.0.0.2", 600, false, "www.google.com"})
	fmt.Println(l.size)
	fmt.Println(l.list)
	l.Add(servers.Service{"A", "10.0.0.3", 600, false, "www.google.com"})
	fmt.Println(l.size)
	fmt.Println(l.list)
	l.Add(servers.Service{"A", "10.0.0.4", 600, false, "www.google.com"})
	fmt.Println(l.size)
	fmt.Println(l.list)
	l.Set(servers.Service{"A", "10.0.0.4", 600, false, "www.google.com"}, servers.Service{"A", "10.0.0.4", 600, true, "www.google.com"})
	if result := l.Set(servers.Service{"A", "10.0.0.4", 600, false, "www.baidu.com"},
		servers.Service{"A", "10.0.0.4", 600, true, "www.google.com"}); result != false {
		t.Errorf("not get false")
	}
	fmt.Println(l.List(servers.Service{"MX", "", 0, false, "www.taobao.com"}))
	fmt.Println(l.List(servers.Service{"MX", "", 0, false, "www.google.com"}))
	l.Add(servers.Service{"A", "11.0.0.0", 600, true, "www.google.com"})
	l.Add(servers.Service{"MX", "11.0.0.0", 600, true, "www.google.com"})
	fmt.Println("1")
	l.Remove(servers.Service{"A", "11.0.0.0", 600, true, "www.google.com"})
	l.Remove(servers.Service{"MX", "www.baidu.com", 600, true, "www.google.com"})
	fmt.Println("2")
	l.Add(servers.Service{"MX", "11.0.0.0", 600, true, "www.google.com"})
	fmt.Println("3")
	l.Add(servers.Service{"MX", "12.0.0.0", 600, true, "www.google.com"})
	fmt.Println("4")
	l.Add(servers.Service{"MX", "13.0.0.0", 600, true, "www.google.com"})
	fmt.Println("5")
	l.Clear()
}
