package simplemsglru

import (
	"fmt"
	"github.com/miekg/dns"
	"testing"
)

func getmsg(name string, rtype uint16) dns.Msg {
	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{name, rtype, dns.ClassINET}
	c := new(dns.Client)
	in, _, _ := c.Exchange(m1, "8.8.8.8:53")
	return *in
}

func TestSimleMsgLRU(t *testing.T) {
	_, err := NewLRU(0, func(s *[]dns.RR) { fmt.Println(*s) })
	if err == nil {
		t.Errorf("should get a error")
	}
	l, err := NewLRU(3, func(s *[]dns.RR) { fmt.Println(*s) })
	if err != nil {
		t.Errorf("fail to create LRU")
	}
	l.Add(getmsg("www.baidu.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.google.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.google.com.", dns.TypeAAAA).Answer, dns.TypeAAAA)
	l.Add(getmsg("www.renren.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.taobao.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.weibo.com.", dns.TypeA).Answer, dns.TypeA)
	fmt.Println(l.Get("www.baidu.com.", dns.TypeA))
	fmt.Println(l.Get("www.weibo.com.", dns.TypeA))
	fmt.Println(l.Get("www.weibo.com.", dns.TypeCNAME))
	fmt.Println(l.Len())
	fmt.Println(l.Keys())
	l.Contains("www.baidu.com.")
	l.Remove("www.weibo.com.", dns.TypeA)
	l.Remove("www.weibo.com.", dns.TypeCNAME)
	l.Remove("www.wwweeeiiiibo.com.", dns.TypeA)

	l.Purge()
}
