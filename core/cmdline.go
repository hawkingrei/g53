package core

import (
	"fmt"
	"strconv"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/hawkingrei/G53/utils"
)

const (
	// VERSION G53 version
	VERSION = "0.0.1"
)

// CommandLine structure handling parameter parsing
type CommandLine struct{}

// ParseParameters Parse parameters
func (cmdline *CommandLine) ParseParameters(rawParams []string) (res *utils.Config, err error) {
	res = utils.NewConfig()

	app := kingpin.New("G53", "Automatic DNS.")
	app.Version(VERSION)
	app.HelpFlag.Short('h')

	nameservers := app.Flag("nameserver", "Comma separated list of DNS server(s) for unmatched requests").Default("114.114.114.114:53").Strings()
	dns := app.Flag("dns", "Listen DNS requests on this address").Default(res.DnsAddr).Short('d').String()
	http := app.Flag("http", "Listen HTTP requests on this address").Default(res.HttpAddr).Short('t').String()
	domain := app.Flag("domain", "Domain that is appended to all requests").Default(res.Domain.String()).String()
	environment := app.Flag("environment", "Optional context before domain suffix").Default("").String()
	ttl := app.Flag("ttl", "TTL for matched requests").Default(strconv.FormatInt(int64(res.Ttl), 10)).Int()
	createAlias := app.Flag("alias", "Automatically create an alias with just the container name.").Default(strconv.FormatBool(res.CreateAlias)).Bool()
	verbose := app.Flag("verbose", "Verbose mode.").Default(strconv.FormatBool(res.Verbose)).Short('v').Bool()
	quiet := app.Flag("quiet", "Quiet mode.").Default(strconv.FormatBool(res.Quiet)).Short('q').Bool()

	kingpin.MustParse(app.Parse(rawParams))

	res.Verbose = *verbose
	res.Quiet = *quiet
	res.Nameservers = *nameservers
	res.DnsAddr = *dns
	res.HttpAddr = *http
	res.Domain = utils.NewDomain(fmt.Sprintf("%s.%s", *environment, *domain))
	res.Ttl = *ttl
	res.CreateAlias = *createAlias
	return
}
