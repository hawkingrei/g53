package cmdline

import (
	"bytes"
	"gopkg.in/alecthomas/kingpin.v2"
	"runtime"
	"strconv"
	"text/template"

	"github.com/hawkingrei/g53/utils"
	"github.com/hawkingrei/g53/version"
)

// CommandLine structure handling parameter parsing
type CommandLine struct{}

var versionTemplate = `Client:
 Version:      {{.Version}}
 Go version:   {{.GoVersion}}
 Git commit:   {{.GitCommit}}
 Built:        {{.BuildTime}}
 OS/Arch:      {{.Os}}/{{.Arch}}`

// ParseParameters Parse parameters
func (cmdline *CommandLine) ParseParameters(rawParams []string) (res *utils.Config, err error) {
	var doc bytes.Buffer
	res = utils.NewConfig()

	vo := version.VersionOptions{
		GitCommit: version.GitCommit,
		Version:   version.Version,
		BuildTime: version.BuildTime,
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

	nameservers := app.Flag("nameserver", "Comma separated list of DNS server(s) for unmatched requests").Default("8.8.8.8:53,8.8.4.4:53").Strings()
	dns := app.Flag("dns", "Listen DNS requests on this address").Default(res.DnsAddr).Short('d').String()
	http := app.Flag("http", "Listen HTTP requests on this address").Default(res.HttpAddr).Default(":80").String()
	ttl := app.Flag("ttl", "TTL for matched requests").Default(strconv.FormatInt(int64(res.Ttl), 10)).Int()
	
	verbose := app.Flag("verbose", "Verbose mode.").Default(strconv.FormatBool(res.Verbose)).Short('v').Bool()
	quiet := app.Flag("quiet", "Quiet mode.").Default(strconv.FormatBool(res.Quiet)).Short('q').Bool()

	kingpin.MustParse(app.Parse(rawParams))
	res.Verbose = *verbose
	res.Quiet = *quiet
	res.Nameservers = *nameservers
	res.DnsAddr = *dns
	res.HttpAddr = *http
	res.Ttl = *ttl
	return
}
