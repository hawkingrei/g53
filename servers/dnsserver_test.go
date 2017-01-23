package servers

import (
	"github.com/hawkingrei/G53/utils"
	"github.com/miekg/dns"
	"testing"
	"time"
)

func TestDNSError(t *testing.T) {
	const TestAddr = "127.0.0.1:9953"

	config := utils.NewConfig()
	config.DnsAddr = TestAddr
	config.Nameservers = []string{"180.76.76.77:53"}

	server := NewDNSServer(config)
	go server.Start()

	// Allow some time for server to start
	time.Sleep(250 * time.Millisecond)

	var inputs = []struct {
		query    string
		expected int
		qType    string
		rcode    int
	}{
		{"google.com.", -1, "A", dns.RcodeRefused},
	}
	c := new(dns.Client)
	c.Timeout = 15 * time.Second
	c.DialTimeout = 15 * time.Second
	c.ReadTimeout = 15 * time.Second
	c.WriteTimeout = 15 * time.Second
	c.UDPSize = uint16(4096)
	for _, input := range inputs {
		qType := dns.StringToType[input.qType]

		m := new(dns.Msg)
		m.SetEdns0(4096, true)
		m.SetQuestion(input.query, qType)
		in, _, err := c.Exchange(m, TestAddr)
		//t.Log(r)
		if err != nil {
			t.Error("Error response from the server", err)
			break
		}
		if in.Rcode != input.rcode {
			t.Error(input, "Rcode expected:",
				dns.RcodeToString[input.rcode],
				" got:", dns.RcodeToString[in.Rcode])
		}
	}
	server.Stop()
	time.Sleep(250 * time.Millisecond)
}

func TestDNSResponse(t *testing.T) {
	const TestAddr = "127.0.0.1:9954"

	config := utils.NewConfig()
	config.DnsAddr = TestAddr

	server := NewDNSServer(config)
	go server.Start()

	// Allow some time for server to start
	time.Sleep(250 * time.Millisecond)

	//server.AddService("www.duitang.net", Service{RecordType: "CNAME", TTL: 600 , Value: "www.cctv.com",Aliases: "www.duitang.net"})
	server.AddService("a.duitang.net", Service{RecordType: "A", TTL: 600, Value: "127.0.0.1", Aliases: "a.duitang.net"})
	server.AddService("b.duitang.net", Service{RecordType: "CNAME", TTL: 600, Value: "wiki.duitang.com", Aliases: "b.duitang.net"})
	//server.AddService("b.duitang.net", Service{RecordType:"MX",TTL:60,Value:"mxbiz1.qq.com.",Aliases:"b.duitang.net"})
	//server.AddService("foo", Service{Name: "foo", Image: "bar", IPs: []net.IP{net.ParseIP("127.0.0.1")}})
	//server.AddService("baz", Service{Name: "baz", Image: "bar", IPs: []net.IP{net.ParseIP("127.0.0.1")}, TTL: -1})
	//server.AddService("biz", Service{Name: "hey", Image: "", IPs: []net.IP{net.ParseIP("127.0.0.4")}})
	//server.AddService("joe", Service{Name: "joe", Image: "", IPs: []net.IP{net.ParseIP("127.0.0.5")}, Aliases: []string{"lala.docker", "super-alias", "alias.domain"}})

	var inputs = []struct {
		query    string
		expected int
		qType    string
		rcode    int
	}{
		{"google.com.", -1, "A", dns.RcodeSuccess},
		{"google.com.", -1, "AAAA", dns.RcodeSuccess}, // baidu has AAAA records
		{"google.com.", -1, "MX", dns.RcodeSuccess},
		{"wiki.duitang.net.", -1, "CNAME", dns.RcodeSuccess},
		{"a.duitang.net.", -1, "A", dns.RcodeSuccess},
		{"a.duitang.net.", -1, "A", dns.RcodeSuccess},
		{"b.duitang.net.", -1, "CNAME", dns.RcodeSuccess},
		{"foo.docker.", 0, "A", dns.RcodeNameError},
		//{"www.1d1u1i1t1a1n1g1.net.",-1,"A",dns.RcodeSuccess},
		//{"docker.", 5, "A", 0},
		//{"docker.", 5, "MX", 0},
		//{"*.docker.", 5, "A", 0},
		//{"*.docker.", 5, "MX", 0},
		//{"bar.docker.", 2, "A", 0},
		//{"bar.docker.", 2, "MX", 0},
		//{"bar.docker.", 0, "AAAA", 0},
		//{"foo.docker.", 0, "A", dns.RcodeNameError},
		//{"foo.docker.", 0, "MX", dns.RcodeNameError},
		//{"baz.bar.docker.", 1, "A", 0},
		//{"baz.bar.docker.", 1, "MX", 0},
		//{"joe.docker.", 1, "A", 0},
		//{"joe.docker.", 1, "MX", 0},
		//{"joe.docker.", 0, "AAAA", 0},
		//{"super-alias.", 1, "A", 0},
		//{"super-alias.", 1, "MX", 0},
		//{"alias.domain.", 1, "A", 0},
		//{"alias.domain.", 1, "MX", 0},
		//{"1.0.0.127.in-addr.arpa.", 4, "PTR", 0},                  // two services match with two domains each
		//{"5.0.0.127.in-addr.arpa.", 4, "PTR", 0},                  // one service match with three aliases
		//{"4.0.0.127.in-addr.arpa.", 1, "PTR", 0},                  // only one service with a single domain
		//{"2.0.0.127.in-addr.arpa.", 0, "PTR", dns.RcodeNameError}, // no match
	}

	c := new(dns.Client)
	c.Timeout = 15 * time.Second
	c.DialTimeout = 15 * time.Second
	c.ReadTimeout = 15 * time.Second
	c.WriteTimeout = 15 * time.Second
	c.UDPSize = uint16(4096)
	for _, input := range inputs {
		t.Log("Query", input.query, input.qType)
		qType := dns.StringToType[input.qType]

		m := new(dns.Msg)
		m.SetEdns0(4096, true)
		m.SetQuestion(input.query, qType)
		r, _, err := c.Exchange(m, TestAddr)

		if err != nil {
			t.Error("Error response from the server", err)
			break
		}

		if input.expected > 0 && len(r.Answer) != input.expected {
			t.Error(input, "Expected:", input.expected,
				" Got:", len(r.Answer))
		}

		if input.expected < 0 && len(r.Answer) == 0 {
			t.Error(input, "Expected at least one record but got none")
		}

		if r.Rcode != input.rcode {
			t.Error(input, "Rcode expected:",
				dns.RcodeToString[input.rcode],
				" got:", dns.RcodeToString[r.Rcode])
		}

		for _, a := range r.Answer {
			rrType := dns.Type(a.Header().Rrtype).String()
			if input.qType != rrType {
				t.Error("Did not receive ", input.qType, " resource record")
			} else {
				t.Log("Received expected response RR type", rrType, "code", dns.RcodeToString[input.rcode])
			}
		}
	}
	server.Stop()
}

func TestServiceManagement(t *testing.T) {
	list := ServiceListProvider(NewDNSServer(utils.NewConfig()))

	if len(list.GetAllServices()) != 0 {
		t.Error("Initial service count should be 0.")
	}

	A := Service{Aliases: "bar.duitang.com.", RecordType: "A", TTL: 3600, Value: "127.0.0.1"}
	list.AddService("bar.duitang.com.", A)

	if len(list.GetAllServices()) != 1 {
		t.Error("Service count should be 1.")
	}

	A.Aliases = "baz.duitang.com."

	s1, err := list.GetService("bar.duitang.com.")
	if err != nil {
		t.Error("GetService error", err)
	}

	if s1.Aliases != "bar.duitang.com." {
		t.Error("Expected: bar got:", s1.Aliases)
	}

	_, err = list.GetService("boo.duitang.com.")
	if err == nil {
		t.Error("Request to boo should have failed")
	}

	list.AddService("boo.duitang.com.", Service{Aliases: "boo.duitang.com.", TTL: 3600, RecordType: "A", Value: "127.0.0.1"})

	all := list.GetAllServices()

	delete(all, "bar.duitang.com.")
	s2 := all["boo.duitang.com."]
	s2.Aliases = "zoo.duitang.com."

	if len(list.GetAllServices()) != 2 {
		t.Error("Local map change should not remove items")
	}

	if s1, _ = list.GetService("boo.duitang.com."); s1.Aliases != "boo.duitang.com." {
		t.Error("Local map change should not change items")
	}

	err = list.RemoveService("barr.duitang.com.")
	if err == nil {
		t.Error("Removing bar.duitang.com. should fail")
	}

	err = list.RemoveService("boo.duitang.com.")
	if err != nil {
		t.Error("Removing boo.duitang.com. failed", err)
	}

	if len(list.GetAllServices()) != 1 {
		t.Log(len(list.GetAllServices()))
		t.Error("Item count after remove should be 1")
	}

	list.AddService("416261e74515b7dd1dbd55f35e8625b063044f6ddf74907269e07e9f142bc0df", Service{Aliases: "mysql.duitang.net.", RecordType: "A", Value: "127.0.0.1"})

	if s1, _ = list.GetService("416261"); s1.Aliases != "mysql.duitang.net." {
		t.Error("Container can't be found by prefix")
	}

	err = list.RemoveService("416261")
	if err != nil {
		t.Error("Removing 416261 failed", err)
	}

	if len(list.GetAllServices()) != 1 {
		t.Error("Item count after remove should be 1")
	}
}

func TestGetExpandedID(t *testing.T) {
	server := NewDNSServer(utils.NewConfig())

	server.AddService("416261e74515b7dd1dbd55f35e8625b063044f6ddf74907269e07e9f142bc0df", Service{RecordType: "A", TTL: 600, Value: "127.0.0.1", Aliases: "www.416261e74.net."})
	server.AddService("316261e74515b7dd1dbd55f35e8625b063044f6ddf74907269e07e9f14nothex", Service{RecordType: "A", TTL: 600, Value: "127.0.0.1", Aliases: "www.316261e74.net."})
	server.AddService("abcdefabcdef", Service{RecordType: "A", TTL: 600, Value: "127.0.0.1", Aliases: "www.abcdefabcdef.net."})

	inputs := map[string]string{
		"416":          "416",
		"41626":        "416261e74515b7dd1dbd55f35e8625b063044f6ddf74907269e07e9f142bc0df",
		"416261e74515": "416261e74515b7dd1dbd55f35e8625b063044f6ddf74907269e07e9f142bc0df",
		"31626":        "31626",
		"abcde":        "abcde",
		"foobar":       "foobar",
	}

	for input, expected := range inputs {
		if actual := server.getExpandedID(input); actual != expected {
			t.Error(input, "Expected:", expected, "Got:", actual)
		}
	}

}
