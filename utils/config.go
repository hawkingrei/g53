package utils

import (
	"strings"
)

// Domain represents a domain
type Domain []string

// NewDomain creates a new domain
func NewDomain(s string) Domain {
	s = strings.Replace(s, "..", ".", -1)
	if s[:1] == "." {
		s = s[1:]
	}
	if s[len(s)-1:] == "." {
		s = s[:len(s)-1]
	}
	return Domain(strings.Split(s, "."))
}

func (d *Domain) String() string {
	return strings.Join([]string(*d), ".")
}

// type that knows how to parse CSV strings and store the values in a slice
type nameservers []string

func (n *nameservers) String() string {
	return strings.Join(*n, " ")
}

// accumulate the CSV string of nameservers
func (n *nameservers) Set(value string) error {
	*n = nil
	for _, ns := range strings.Split(value, ",") {
		ns = strings.Trim(ns, " ")
		*n = append(*n, ns)
	}

	return nil
}

// Config contains DNSDock configuration
type Config struct {
	Nameservers nameservers
	DnsAddr     string
	Domain      Domain
	TlsVerify   bool
	TlsCaCert   string
	TlsCert     string
	TlsKey      string
	HttpAddr    string
	Ttl         int
	CreateAlias bool
	Verbose     bool
	Quiet       bool
}

// NewConfig creates a new config
func NewConfig() *Config {
	return &Config{
		Nameservers: nameservers{"8.8.4.4:53", "8.8.8.8:53"},
		DnsAddr:     ":53",
		Domain:      NewDomain("suphawking.com"),
		//DockerHost:  dockerHost,
		HttpAddr:    ":80",
		CreateAlias: false,
		/*
			TlsVerify:   tlsVerify,
			TlsCaCert:   dockerCerts + "/ca.pem",
			TlsCert:     dockerCerts + "/cert.pem",
			TlsKey:      dockerCerts + "/key.pem",
		*/
		Verbose: false,
		Quiet:   false,
	}

}
