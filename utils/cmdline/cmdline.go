package cmdline

import (
	"bytes"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"runtime"
	"strconv"
	"text/template"

	"github.com/hawkingrei/g53/utils"
)

// CommandLine structure handling parameter parsing
type CommandLine struct{}

var versionTemplate = `Client:
 Version:      {{.Version}}
 Go version:   {{.GoVersion}}
 Git commit:   {{.GitCommit}}
 Built:        {{.BuildTime}}
 OS/Arch:      {{.Os}}/{{.Arch}}`

type VersionOptions struct {
	GitCommit string
	Version   string
	BuildTime string
	GoVersion string
	Os        string
	Arch      string
}


// ParseParameters Parse parameters
func (cmdline *CommandLine) ParseParameters(rawParams []string) (res *utils.Config, err error) {
	var doc bytes.Buffer
	res = utils.NewConfig()

	vo := VersionOptions{
		GitCommit: GitCommit,
		Version:   Version,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
	templateFormat := versionTemplate
	tmpl, _ := template.New("version").Parse(templateFormat)
	tmpl.Execute(&doc, vo)
	VERSION := doc.String()
	app := kingpin.New("G53", "Automatic DNS.")
	app.Version(VERSION)
	app.HelpFlag.Short('h')

	nameservers := app.Flag("nameserver", "Comma separated list of DNS server(s) for unmatched requests").Default("114.114.114.114:53").Strings()
	dns := app.Flag("dns", "Listen DNS requests on this address").Default(res.DnsAddr).Short('d').String()
	http := app.Flag("http", "Listen HTTP requests on this address").Default(res.HttpAddr).Default(":80").String()
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
