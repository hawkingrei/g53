package cache

import (
	"fmt"
	"github.com/hawkingrei/g53/utils"
	"testing"
)

func TestLRU(t *testing.T) {
	_, err := NewWithEvict(0, func(s *utils.Entry) { fmt.Println(*s) })
	if err == nil {
		t.Errorf("should get a error")
	}
	l, err := NewWithEvict(3, func(s *utils.Entry) { fmt.Println(*s) })
	if err != nil {
		t.Errorf("fail to create LRU")
	}
	l.Add(utils.Service{"A", "10.0.0.0", 600, "www.google.com"})
	l.Get(utils.Service{"A", "", 0, "www.google.com"})
	l.Remove(utils.Service{"A", "", 0, "www.google.com"})
	l.Get(utils.Service{"A", "", 0, "www.google.com"})
	if tmp, _ := l.Get(utils.Service{"MX", "", 0, "www.google.com"}); (*tmp != utils.Entry{}) {
		t.Errorf("not get nil")
	}
	if tmp, _ := l.Get(utils.Service{"MX", "", 0, "www.taobao.com"}); (*tmp != utils.Entry{}) {
		t.Errorf("not get nil")
	}
	fmt.Println(l.Keys())
	fmt.Println(l.Len())
	l.Purge()
	fmt.Println(l.Keys())
	fmt.Println(l.Len())
	l.Add(utils.Service{"A", "11.0.0.0", 600, "www.google.com"})
	l.Add(utils.Service{"MX", "11.0.0.0", 600, "www.google.com"})
	l.Remove(utils.Service{"A", "11.0.0.0", 600, "www.google.com"})
	l.Remove(utils.Service{"MX", "www.baidu.com", 600, "www.google.com"})
	l.Add(utils.Service{"MX", "11.0.0.0", 600, "www.google.com"})
	l.Add(utils.Service{"MX", "12.0.0.0", 600, "www.google.com"})
	l.Add(utils.Service{"MX", "13.0.0.0", 600, "www.google.com"})
	fmt.Println(l.Keys())
	fmt.Println(l.Len())
	l.Set(utils.Service{"A", "10.0.0.0", 500, "www.google.com"}, utils.Service{"A", "10.0.0.1", 500, "www.google.com"})
	l.Add(utils.Service{"A", "10.0.0.2", 600, "www.google.com"})
	l.Add(utils.Service{"A", "10.0.0.3", 600, "www.google.com"})
	l.Add(utils.Service{"A", "10.0.0.4", 600, "www.google.com"})
	if result := l.Set(utils.Service{"A", "10.0.0.4", 600, "www.renren.com"},
		utils.Service{"A", "10.0.0.5", 600, "www.google.com"}); result == nil {
		t.Errorf("not get nil")
	}
	if result := l.Set(utils.Service{"MX", "10.0.0.4", 600, "www.google.com"},
		utils.Service{"MX", "10.0.0.5", 600, "www.google.com"}); result == nil {
		t.Errorf("not get nil")
	}
	if result := l.Set(utils.Service{"A", "10.0.0.10", 600, "www.google.com"},
		utils.Service{"A", "10.0.0.5", 600, "www.google.com"}); result == nil {
		t.Errorf("not get nil")
	}
	l.Add(utils.Service{"A", "11.0.0.0", 600, "www.google.com"})
	l.Add(utils.Service{"MX", "11.0.0.0", 600, "www.google.com"})
	fmt.Println("1")
	l.Remove(utils.Service{"A", "11.0.0.0", 600, "www.google.com"})
	l.Remove(utils.Service{"MX", "www.baidu.com", 600, "www.google.com"})
	fmt.Println("2")
	l.Add(utils.Service{"MX", "11.0.0.0", 600, "www.google.com"})
	fmt.Println("3")
	l.Add(utils.Service{"MX", "12.0.0.0", 600, "www.google.com"})
	fmt.Println("4")
	l.Add(utils.Service{"MX", "13.0.0.0", 600, "www.google.com"})
	fmt.Println("5")
	l.Purge()
	l.Add(utils.Service{"MX", "13.0.0.0", 600, "www.oschina.com"})
	if result := l.Set(utils.Service{"MX", "13.0.0.0", 600, "www.oschina.com"},
		utils.Service{"MX", "13.0.0.1", 600, "www.oschina.com"}); result != nil {
		t.Errorf("should get nil")
	}
	fmt.Println(l.Get(utils.Service{"MX", "", 600, "www.oschina.com"}))
	l.RemoveOldest()
	_, err = New(3)
	if err != nil {
		t.Errorf("fail to create LRU")
	}
}
