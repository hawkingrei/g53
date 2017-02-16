package servers

import (
	"errors"
	"net"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"

	"github.com/hawkingrei/g53/utils"
)

// Service represents a container and an attached DNS record
// service(recode_type: "A",value: []string{"127.0.0.1","127.0.0.1"},Aliases: "www.duitang.net" ))
type Service struct {
	RecordType string
	Value      string
	TTL        int
	Aliases    string
	private    bool
}

// NewService creates a new service
func NewService() (s *Service) {
	s = &Service{TTL: -1}
	return
}

// ServiceListProvider represents the entrypoint to get containers
type ServiceListProvider interface {
	AddService(string, Service)
	RemoveService(string) error
	GetService(string) (Service, error)
	GetAllServices() map[string]Service
}

// DNSServer represents a DNS server
type DNSServer struct {
	config   *utils.Config
	server   *dns.Server
	mux      *dns.ServeMux
	services map[string]*Service
	lock     *sync.RWMutex
}

// NewDNSServer create a new DNSServer
func NewDNSServer(c *utils.Config) *DNSServer {
	s := &DNSServer{
		config:   c,
		services: make(map[string]*Service),
		lock:     &sync.RWMutex{},
	}

	logger.Debugf("Handling DNS requests for '%s'.", c.Domain.String())

	s.mux = dns.NewServeMux()
	s.mux.HandleFunc(".", s.handleRequest)
	s.server = &dns.Server{Addr: c.DnsAddr, Net: "udp", Handler: s.mux}

	return s
}

// Start starts the DNSServer
func (s *DNSServer) Start() error {
	logger.Infof("start")
	return s.server.ListenAndServe()
}

// Stop stops the DNSServer
func (s *DNSServer) Stop() {
	s.server.Shutdown()
}

// AddService adds a new container and thus new DNS records
func (s *DNSServer) AddService(id string, service Service) {
	if service.RecordType == "CNAME" || service.RecordType == "A" {
		defer s.lock.Unlock()
		s.lock.Lock()

		id = s.getExpandedID(id)
		s.services[id] = &service

		logger.Debugf("Added service: '%s': '%s'.", id, service)
		logger.Debugf("Handling DNS requests for '%s'.", service.Aliases)
		s.mux.HandleFunc(service.Aliases+".", s.handleRequest)
		//for _, alias := range service.Aliases {
		//	logger.Debugf("Handling DNS requests for '%s'.", alias)
		//	s.mux.HandleFunc(alias+".", s.handleRequest)
		//}
	} else {
		logger.Warningf("Service '%s' ignored: No RecordType provided:", id, id)
	}
}

// RemoveService removes a new container and thus DNS records
func (s *DNSServer) RemoveService(id string) error {
	defer s.lock.Unlock()
	s.lock.Lock()

	id = s.getExpandedID(id)
	if _, ok := s.services[id]; !ok {
		return errors.New("No such service: " + id)
	}
	s.mux.HandleRemove(s.services[id].Aliases + ".")

	delete(s.services, id)

	logger.Debugf("Removeed service '%s'", id)

	return nil
}

// GetService reads a service from the repository
func (s *DNSServer) GetService(id string) (Service, error) {
	defer s.lock.RUnlock()
	s.lock.RLock()

	id = s.getExpandedID(id)
	if s, ok := s.services[id]; ok {
		return *s, nil
	}
	// Check for a pa
	return *new(Service), errors.New("No such service: " + id)
}

// GetAllServices reads all services from the repository
func (s *DNSServer) GetAllServices() map[string]Service {
	defer s.lock.RUnlock()
	s.lock.RLock()

	list := make(map[string]Service, len(s.services))
	for id, service := range s.services {
		list[id] = *service
	}

	return list
}

func (s *DNSServer) handleForward(w dns.ResponseWriter, r *dns.Msg) {
	//r.SetEdns0(4096, true)
	logger.Debugf("Using DNS forwarding for '%s'", r.Question[0].Name)
	logger.Debugf("Forwarding DNS nameservers: %s", s.config.Nameservers.String())

	// Otherwise just forward the request to another server
	c := new(dns.Client)
	c.UDPSize = uint16(4096)

	// look at each Nameserver, stop on success
	for i := range s.config.Nameservers {
		logger.Debugf("Using Nameserver %s", s.config.Nameservers[i])

		in, _, err := c.Exchange(r, s.config.Nameservers[i])
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

func (s *DNSServer) makeServiceCNAME(n string, service *Service) dns.RR {
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

	if len(service.Value) != 0 {
		if len(service.Value) > 1 {
			logger.Warningf("Multiple IP address found for container '%s'. Only the first address will be used", service.Aliases)
		}
		rr.Target = service.Value + "."
	} else {
		logger.Errorf("No valid IP address found for container '%s' ", service.Aliases)
	}
	return rr
}

func (s *DNSServer) makeServiceA(n string, service *Service) dns.RR {
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

	if len(service.Value) != 0 {
		if len(service.Value) > 1 {
			logger.Warningf("Multiple IP address found for container '%s'. Only the first address will be used", service.Aliases)
		}
		rr.A = net.ParseIP(service.Value)
	} else {
		logger.Errorf("No valid IP address found for container '%s' ", service.Aliases)
	}

	return rr
}

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
	if query[len(query)-1] == '.' {
		query = query[:len(query)-1]
	}

	logger.Debugf("DNS request for query '%s' from remote '%s'", w.RemoteAddr().String(), w.RemoteAddr())
	result := s.queryServices(query)
	logger.Debugf("DNS record found for query '%s'", query)
	for service := range result {
		var rr dns.RR
		switch r.Question[0].Qtype {
		case dns.TypeA:
			rr = s.makeServiceA(r.Question[0].Name, service)
		case dns.TypeCNAME:
			rr = s.makeServiceCNAME(r.Question[0].Name, service)
		default:
			// this query type isn't supported, but we do have
			// a record with this name. Per RFC 4074 sec. 3, we
			// immediately return an empty NOERROR reply.
			m.Ns = s.createSOA()
			m.MsgHdr.Authoritative = true
			w.WriteMsg(m)
			return
		}

		m.Answer = append(m.Answer, rr)
	}

	// We didn't find a record corresponding to the query
	if len(m.Answer) == 0 {

		s.handleForward(w, r)
		return
	}
	w.WriteMsg(m)
	return
}

func (s *DNSServer) queryServices(query string) chan *Service {
	c := make(chan *Service, 10)

	go func() {
		query := strings.Split(strings.ToLower(query), ".")

		defer s.lock.RUnlock()
		s.lock.RLock()

		for _, service := range s.services {
			// create the name for this service, skip empty strings
			test := []string{}
			// todo: add some cache to avoid calculating this every time

			test = append(test, strings.Split(service.Aliases, ".")...)

			if isPrefixQuery(query, test) {
				c <- service
			}
		}
		close(c)
	}()
	return c
}

// Checks for a partial match for container SHA and outputs it if found.
func (s *DNSServer) getExpandedID(in string) (out string) {
	out = in

	// Hard to make a judgement on small image names.
	if len(in) < 4 {
		return
	}

	if isHex, _ := regexp.MatchString("^[0-9a-f]+$", in); !isHex {
		return
	}

	for id := range s.services {
		if len(id) == 64 {
			if isHex, _ := regexp.MatchString("^[0-9a-f]+$", id); isHex {
				if strings.HasPrefix(id, in) {
					out = id
					return
				}
			}
		}
	}
	return
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

func isPrefixQuery(query, name []string) bool {
	return reflect.DeepEqual(query, name)
}
