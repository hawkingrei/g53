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
		if recordType == dns.TypeA || recordType == dns.TypeAAAA {
			result, err = s.Get(name, dns.TypeCNAME)
			for v := 0; v < len(result); v++ {
				m.Answer = append(m.Answer, result[v])
			}
			if err != nil {
				return m, err
			}
		}
		return m, err
	}
	if recordType == dns.TypeCNAME {
		for v := 0; v < len(result); v++ {
			if result[v].Header().Rrtype == dns.TypeCNAME {
				m.Answer = append(m.Answer, result[v])
				continue
			}
			m.Extra = append(m.Extra, result[v])
		}
		return m, nil
	}

	for v := 0; v < len(result); v++ {
		m.Answer = append(m.Answer, result[v])
	}

	return m, nil
}
