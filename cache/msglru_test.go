package cache

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

func TestMsgLRU(t *testing.T) {
	_, err := NewMsgCacheWithEvict(0, func(s *[]dns.RR) { fmt.Println(*s) })
	if err == nil {
		t.Errorf("should get a error")
	}
	l, err := NewMsgCacheWithEvict(256*1, func(s *[]dns.RR) { fmt.Println(*s) })
	if err != nil {
		t.Errorf("fail to create LRU")
	}
	l.Add(getmsg("www.baidu.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.renren.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.taobao.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.weibo.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.qq.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.sohu.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.tmall.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.jd.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.hao123.com.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.csdn.net.", dns.TypeA).Answer, dns.TypeA)
	l.Add(getmsg("www.soso.com.", dns.TypeA).Answer, dns.TypeA)

	fmt.Println(l.Get("www.baidu.com.", dns.TypeA))
	fmt.Println(l.Get("www.weibo.com.", dns.TypeA))
	fmt.Println(l.Len())
	fmt.Println(l.Keys())
	l.Purge()
}
