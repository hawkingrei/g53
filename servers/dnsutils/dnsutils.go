package dnsutils

import (
	"errors"
	"time"
	"github.com/hawkingrei/g53/cache"
	"github.com/miekg/dns"
)

func Round(val float64) uint32 {
	if val < 0 {
		return uint32(val - 0.5)
	}
	return uint32(val + 0.5)
}

func QueryDnsCache(s *cache.MsgCache, r *dns.Msg) (*dns.Msg, error) {
	m := new(dns.Msg)
	m.Compress = true
	m.SetReply(r)
	m.RecursionAvailable = true

	name := r.Question[0].Name
	recordType := r.Question[0].Qtype
	result, rtime, err := s.Get(name, recordType)
	if err != nil {
		return m, err
	}
	nowtime := time.Now()
	var rr []dns.RR = make([]dns.RR, len(result))
	for v := 0; v < len(result); v++ {
		expiration := rtime.Add(time.Duration(result[v].Header().Ttl) * time.Second)
		if expiration.Before(nowtime) {
			s.Remove(name, recordType)
			return m, errors.New("expiration")
		}
	}
	cttl := nowtime.Sub(*rtime).Seconds()
	for v := 0; v < len(result); v++ {
		rr[v] = dns.Copy(result[v])
	}
	for v := 0; v < len(rr); v++ {
		tmp := rr[v].Header().Ttl - Round(cttl)
		if tmp <= 1 {
			s.Remove(name, recordType)
			return m, errors.New("expiration")
		}
		rr[v].Header().Ttl = tmp
	}
	result = rr
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

	if recordType == dns.TypeA || recordType == dns.TypeAAAA {
		if result[len(result)-1].Header().Rrtype != dns.TypeA || result[len(result)-1].Header().Rrtype != dns.TypeAAAA && result[len(result)-1].Header().Rrtype != dns.TypeSOA {
			s.Remove(name, recordType)
			return m, nil
		}
	}

	for v := 0; v < len(result); v++ {
		if result[v].Header().Rrtype != dns.TypeSOA {
			m.Answer = append(m.Answer, result[v])
			continue
		}
		m.Extra = append(m.Extra, result[v])
	}

	return m, nil
}
