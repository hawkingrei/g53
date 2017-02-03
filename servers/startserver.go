package servers

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/op/go-logging"

	"github.com/hawkingrei/G53/core"
	"github.com/hawkingrei/G53/utils"
)

func StartServer(rawParams []string) {
	var logger = logging.MustGetLogger("G53.main")
	var cmdLine core.CommandLine
	config, err := cmdLine.ParseParameters(rawParams)
	if err != nil {
		logger.Fatalf(err.Error())
	}
	verbosity := 0
	if config.Quiet == false {
		if config.Verbose == false {
			verbosity = 1
		} else {
			verbosity = 2
		}
	}
	err = utils.InitLoggers(verbosity)
	if err != nil {
		logger.Fatalf("Unable to initialize loggers! %s", err.Error())
	}

	dnsServer := NewDNSServer(config)

	var tlsConfig *tls.Config
	if config.TlsVerify {
		clientCert, err := tls.LoadX509KeyPair(config.TlsCert, config.TlsKey)
		if err != nil {
			logger.Fatalf("Error: '%s'", err)
		}
		tlsConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{clientCert},
		}
		pemData, err := ioutil.ReadFile(config.TlsCaCert)
		if err == nil {
			rootCert := x509.NewCertPool()
			rootCert.AppendCertsFromPEM(pemData)
			tlsConfig.RootCAs = rootCert
		} else {
			logger.Fatalf("Error: '%s'", err)
		}
	}
	logger.Infof("ok dns config")
	logger.Infof(config.DnsAddr)
	httpServer := NewHTTPServer(config, dnsServer)
	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Fatalf("Error: '%s'", err)
		}
	}()

	if err := dnsServer.Start(); err != nil {
		logger.Fatalf("Error: '%s'", err)
	}
}
