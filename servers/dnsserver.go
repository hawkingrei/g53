package servers

import (
	"github.com/miekg/dns"
	"net"
	"strings"
	"time"

	"github.com/hawkingrei/g53/cache"
	"github.com/hawkingrei/g53/servers/dnsutils"
	"github.com/hawkingrei/g53/utils"
)

// NewService creates a new service
func NewService() (s *utils.Service) {
	s = &utils.Service{TTL: -1}
	return
}

// ServiceListProvider represents the entrypoint to get containers
type ServiceListProvider interface {
	AddService(utils.Service)
	RemoveService(utils.Service) error
	//SetService(utils.Service, utils.Service) error
	GetService(utils.Service) ([]utils.Service, error)
	GetAllServices() []utils.Service
}

// DNSServer represents a DNS server
type DNSServer struct {
	config     *utils.Config
	server     *dns.Server
	mux        *dns.ServeMux
	publicDns  *cache.MsgCache
	privateDns *cache.Cache
	dnsclient  *dns.Client
}

// NewDNSServer create a new DNSServer
func NewDNSServer(c *utils.Config) *DNSServer {
	publicDns, _ := cache.NewMsgCache(256 * 1)
	privateDns, _ := cache.New(10000)
	dnsclient := new(dns.Client)
	dnsclient.UDPSize = uint16(4096)
	dnsclient.Timeout = time.Duration(5) * time.Second
	s := &DNSServer{
		config:     c,
		publicDns:  publicDns,
		privateDns: privateDns,
		dnsclient:  dnsclient,
	}

	logger.Debugf("Handling DNS requests for '%s'.", c.Domain.String())

	s.mux = dns.NewServeMux()
	s.mux.HandleFunc(".", s.handleRequest)
	s.server = &dns.Server{Addr: c.DnsAddr, Net: "udp", Handler: s.mux}

	return s
}

// Start starts the DNSServer
func (s *DNSServer) Start() error {
	logger.Infof("start DNS Server")
	return s.server.ListenAndServe()
}

// Stop stops the DNSServer
func (s *DNSServer) Stop() {
	s.server.Shutdown()
}

//func (s *DNSServer) SetService(originalValue utils.Service, modifyValue utils.Service) error {
//	return s.privateDns.Set(originalValue, modifyValue)
//}

// AddService adds a new container and thus new DNS records
func (s *DNSServer) AddService(service utils.Service) {
	if service.RecordType == "CNAME" || service.RecordType == "A" {
		if string(service.Aliases[len(service.Aliases)-1]) != "." {
			service.Aliases = string(service.Aliases) + "."
		}

		if service.RecordType == "CNAME" && string(service.Value[len(service.Value)-1]) != "." {
			service.Value = string(service.Value) + "."
		}

		s.privateDns.Add(service)

		logger.Debugf("Added service: '%s'.", service)
		logger.Debugf("Handling DNS requests for '%s'.", service.Aliases)
		s.mux.HandleFunc(service.Aliases+".", s.handleRequest)
		//for _, alias := range service.Aliases {
		//	logger.Debugf("Handling DNS requests for '%s'.", alias)
		//	s.mux.HandleFunc(alias+".", s.handleRequest)
		//}

	} else {
		logger.Warningf("Service '%s' ignored: No RecordType provided:", service)
	}
}

// RemoveService removes a new container and thus DNS records
func (s *DNSServer) RemoveService(service utils.Service) error {
	if err := s.privateDns.Remove(service); err != nil {
		return err
	}
	s.mux.HandleRemove(service.Aliases + ".")
	logger.Debugf("Removeed service '%s'", service)

	return nil
}

// GetService reads a service from the repository
func (s *DNSServer) GetService(service utils.Service) ([]utils.Service, error) {
	result, err := s.privateDns.Get(service)
	if err != nil {
		return *new([]utils.Service), err
	}
	return utils.BatchEntryToServer(&result), err
}

// GetAllServices reads all services from the repository

func (s *DNSServer) GetAllServices() []utils.Service {
	return s.privateDns.List()
}

func (s *DNSServer) queryDnsCache(r *dns.Msg) (*dns.Msg, error) {
	return dnsutils.QueryDnsCache(s.publicDns, r)
}

func (s *DNSServer) DNSExchange(nameservers string, r *dns.Msg) (*dns.Msg, []dns.RR, error) {
	in, _, err := s.dnsclient.Exchange(r, nameservers)
	if err == nil {
		if len(in.Answer) != 0 {
			logger.Debugf(" '%s' '%s' write Cache", r.Question[0].Name, dns.TypeToString[r.Question[0].Qtype])
			result := in.Answer
			if len(in.Extra) != 0 {
				result = append(result, in.Extra...)
			}
			s.publicDns.Add(result, r.Question[0].Qtype)
		}
		return in, in.Answer, err
	}
	return new(dns.Msg), []dns.RR{}, err
}

func (s *DNSServer) handleForward(w dns.ResponseWriter, r *dns.Msg) {
	// Otherwise just forward the request to another server
	if result, err := s.queryDnsCache(r); err == nil {
		logger.Debugf("'%s' '%S' Hit Public Cache", r.Question[0].Name, dns.TypeToString[r.Question[0].Qtype])
		w.WriteMsg(result)
		return
	}
	logger.Debugf("Using DNS forwarding for '%s'", r.Question[0].Name)
	logger.Debugf("Forwarding DNS nameservers: %s", s.config.Nameservers.String())
	// look at each Nameserver, stop on success
	for i := range s.config.Nameservers {
		//logger.Debugf("Using Nameserver %s", s.config.Nameservers[i])

		in, _, err := s.DNSExchange(s.config.Nameservers[i], r)
		if err == nil {
			w.WriteMsg(in)
			return
		}

		if i == (len(s.config.Nameservers) - 1) {
			logger.Noticef("DNS fowarding for '%s' failed: no more nameservers to try", err.Error())

			// Send failure reply
			m := new(dns.Msg)
			m.SetReply(r)
			m.Ns = s.createSOA()
			m.SetRcode(r, dns.RcodeRefused) // REFUSED
			w.WriteMsg(m)

		} else {
			logger.Errorf("DNS fowarding for '%s' failed: trying next Nameserver...", err.Error())
		}
	}
}

func (s *DNSServer) makeServiceCNAME(n string, service utils.Service) dns.RR {
	rr := new(dns.CNAME)
	var ttl int
	if service.TTL != -1 {
		ttl = service.TTL
	} else {
		ttl = s.config.Ttl
	}

	rr.Hdr = dns.RR_Header{
		Name:   n,
		Rrtype: dns.TypeCNAME,
		Class:  dns.ClassINET,
		Ttl:    uint32(ttl),
	}
	rr.Target = service.Value
	return rr
}

func (s *DNSServer) makeServiceA(n string, service utils.Service) dns.RR {
	rr := new(dns.A)
	var ttl int
	if service.TTL != -1 {
		ttl = service.TTL
	} else {
		ttl = s.config.Ttl
	}

	rr.Hdr = dns.RR_Header{
		Name:   n,
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    uint32(ttl),
	}
	rr.A = net.ParseIP(service.Value)
	return rr
}

//func (s *DNSServer) RecursionPrivate(service utils.Service) dns.RR {
//	result, err := s.privateDns.Get(service)//
//}

func (s *DNSServer) MakePrivateRR(query string, qtype uint16, m *dns.Msg) {
	result := s.queryServices(utils.Service{dns.TypeToString[qtype], "", 0, strings.ToLower(query)})
	for services := range result {
		for i := range services {
			var rr dns.RR
			switch qtype {
			case dns.TypeA:
				rr = s.makeServiceA(query, utils.EntryToServer(&services[i]))
			case dns.TypeCNAME:
				rr = s.makeServiceCNAME(query, utils.EntryToServer(&services[i]))
			default:
				// this query type isn't supported, but we do have
				// a record with this name. Per RFC 4074 sec. 3, we
				// immediately return an empty NOERROR reply.
				m.Ns = s.createSOA()
				m.MsgHdr.Authoritative = true
				continue
			}
			m.Answer = append(m.Answer, rr)

		}
	}
}

//handle with dns request
func (s *DNSServer) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.Compress = true
	m.SetReply(r)
	m.RecursionAvailable = true

	// Send empty response for empty requests
	if len(r.Question) == 0 {
		m.Ns = s.createSOA()
		w.WriteMsg(m)
		return
	}

	// respond to SOA requests
	if r.Question[0].Qtype == dns.TypeSOA {
		m.Answer = s.createSOA()
		w.WriteMsg(m)
		return
	}
	m.Answer = make([]dns.RR, 0, 2)
	query := r.Question[0].Name

	// trim off any trailing dot
	if query[len(query)-1] != '.' {
		query = query + "."
	}
	existDomain := s.privateDns.Containkey(query)
	if existDomain {
		logger.Debugf("DNS record found for query '%s'  '%s'", query, dns.TypeToString[r.Question[0].Qtype])
		s.MakePrivateRR(query, r.Question[0].Qtype, m)
		if len(m.Answer) != 0 && m.Answer[0].Header().Rrtype == dns.TypeSOA {
			w.WriteMsg(m)
			return
		}
		if len(m.Answer) == 0 && (r.Question[0].Qtype == dns.TypeA || r.Question[0].Qtype == dns.TypeAAAA) {
			logger.Debugf("DNS record found for query '%s'  '%s'", query, "CNAME")
			s.MakePrivateRR(query, dns.TypeA, m)
		}

		if len(m.Answer) != 0 && m.Answer[len(m.Answer)-1].Header().Rrtype == dns.TypeCNAME {
			tmplong := len(m.Answer)
			s.MakePrivateRR(m.Answer[len(m.Answer)-1].String(), dns.TypeCNAME, m)
			if len(m.Answer) == tmplong {
				askmsg := new(dns.Msg)
				askmsg.Id = dns.Id()
				askmsg.RecursionDesired = true
				askmsg.Question = make([]dns.Question, 1)
				askmsg.Question[0] = dns.Question{m.Answer[len(m.Answer)-1].String(), dns.TypeCNAME, dns.ClassINET}
				for i := range s.config.Nameservers {
					in, _, err := s.DNSExchange(s.config.Nameservers[i], r)
					if err == nil {
						for v := range in.Answer {
							m.Answer = append(m.Answer, in.Answer[v])
						}
						w.WriteMsg(m)
						return
					}
				}
			}
		}
	}
	// We didn't find a record corresponding to the query
	if !(len(m.Answer) > 0) {
		s.handleForward(w, r)
		return
	}
	w.WriteMsg(m)
	return
}

func (s *DNSServer) queryServices(service utils.Service) chan []utils.Entry {
	c := make(chan []utils.Entry, 10)
	go func() {
		result, err := s.privateDns.Get(service)
		if err == nil {
			logger.Debugf("get the number of records: ", len(result))
			c <- result
		} else {
			logger.Debugf(err.Error())
		}
		close(c)
	}()
	return c
}

// TTL is used from config so that not-found result responses are not cached
// for a long time. The other defaults left as is(skydns source) because they
// do not have an use case in this situation.
func (s *DNSServer) createSOA() []dns.RR {
	dom := dns.Fqdn(s.config.Domain.String() + ".")
	soa := &dns.SOA{
		Hdr: dns.RR_Header{
			Name:   dom,
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    uint32(s.config.Ttl)},
		Ns:      "g53." + dom,
		Mbox:    "g53.g53." + dom,
		Serial:  uint32(time.Now().Truncate(time.Hour).Unix()),
		Refresh: 28800,
		Retry:   7200,
		Expire:  604800,
		Minttl:  uint32(s.config.Ttl),
	}
	return []dns.RR{soa}
}
