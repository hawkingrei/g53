package cache

import (
	"fmt"
	"github.com/hawkingrei/g53/servers"
	"testing"
)

func TestLRU(t *testing.T) {
	_, err := NewWithEvict(0, func(s servers.Service) { fmt.Println(s) })
	if err == nil {
		t.Errorf("should get a error")
	}
	l, err := NewWithEvict(3, func(s servers.Service) { fmt.Println(s) })
	if err != nil {
		t.Errorf("fail to create LRU")
	}
	l.Add(servers.Service{"A", "10.0.0.0", 600, false, "www.google.com"})
	l.Get(servers.Service{"A", "", 0, false, "www.google.com"})
	l.Remove(servers.Service{"A", "", 0, false, "www.google.com"})
	l.Get(servers.Service{"A", "", 0, false, "www.google.com"})
	if tmp, _ := l.Get(servers.Service{"MX", "", 0, false, "www.google.com"}); (tmp != servers.Service{}) {
		t.Errorf("not get nil")
	}
	if tmp, _ := l.Get(servers.Service{"MX", "", 0, false, "www.taobao.com"}); (tmp != servers.Service{}) {
		t.Errorf("not get nil")
	}
	fmt.Println(l.Keys())
	fmt.Println(l.Len())
	l.Purge()
	fmt.Println(l.Keys())
	fmt.Println(l.Len())
	l.Add(servers.Service{"A", "11.0.0.0", 600, true, "www.google.com"})
	l.Add(servers.Service{"MX", "11.0.0.0", 600, true, "www.google.com"})
	l.Remove(servers.Service{"A", "11.0.0.0", 600, true, "www.google.com"})
	l.Remove(servers.Service{"MX", "www.baidu.com", 600, true, "www.google.com"})
	l.Add(servers.Service{"MX", "11.0.0.0", 600, true, "www.google.com"})
	l.Add(servers.Service{"MX", "12.0.0.0", 600, true, "www.google.com"})
	l.Add(servers.Service{"MX", "13.0.0.0", 600, true, "www.google.com"})
	fmt.Println(l.Keys())
	fmt.Println(l.Len())
	l.Set(servers.Service{"A", "10.0.0.0", 500, false, "www.google.com"}, servers.Service{"A", "10.0.0.1", 500, false, "www.google.com"})
	l.Add(servers.Service{"A", "10.0.0.2", 600, false, "www.google.com"})
	l.Add(servers.Service{"A", "10.0.0.3", 600, false, "www.google.com"})
	l.Add(servers.Service{"A", "10.0.0.4", 600, false, "www.google.com"})
	if result := l.Set(servers.Service{"A", "10.0.0.4", 600, false, "www.renren.com"},
		servers.Service{"A", "10.0.0.5", 600, false, "www.google.com"}); result == nil {
		t.Errorf("not get nil")
	}
	if result := l.Set(servers.Service{"MX", "10.0.0.4", 600, false, "www.google.com"},
		servers.Service{"MX", "10.0.0.5", 600, false, "www.google.com"}); result == nil {
		t.Errorf("not get nil")
	}
	if result := l.Set(servers.Service{"A", "10.0.0.10", 600, false, "www.google.com"},
		servers.Service{"A", "10.0.0.5", 600, false, "www.google.com"}); result == nil {
		t.Errorf("not get nil")
	}
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
	l.Purge()
	l.RemoveOldest()
	_, err = New(3)
	if err != nil {
		t.Errorf("fail to create LRU")
	}
}
