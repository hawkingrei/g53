package dnsutils

import (
	"github.com/hawkingrei/g53/cache"
	"github.com/miekg/dns"
)

func QueryDnsCache(s *cache.MsgCache, r *dns.Msg) (*dns.Msg, error) {
	m := new(dns.Msg)
	m.Compress = true
	m.SetReply(r)
	m.RecursionAvailable = true

	name := r.Question[0].Name
	recordType := r.Question[0].Qtype
	result, err := s.Get(name, recordType)
	if err != nil {
		result, err = s.Get(name, dns.TypeCNAME)
		if err != nil {
			return m, err
		}
	}
	for v := 0; v < len(result); v++ {
		m.Answer = append(m.Answer, result[v])
	}
	return m, nil
}
