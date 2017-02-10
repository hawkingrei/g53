package servers

import (
	"github.com/op/go-logging"

	"github.com/hawkingrei/g53/utils"
	"github.com/hawkingrei/g53/utils/cmdline"
)

func StartServer(rawParams []string) {
	var logger = logging.MustGetLogger("G53.main")
	var cmdLine cmdline.CommandLine
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
