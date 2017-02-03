package core

import (
	"testing"
        "reflect"	
//        "github.com/hawkingrei/G53/utils"
)

func TestCmdline(t *testing.T) {
	args := []string{`--nameserver="8.8.8.8:53"`}
	var cmdLine CommandLine
	config, err := cmdLine.ParseParameters(args)
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Log(config.Nameservers)
	if reflect.DeepEqual(config.Nameservers , []string{"8.8.8.8:53"})  {
		t.Error("config Nameservers error")
	}
	args = []string{`--dns="127.0.0.1:83"`}
	config,err = cmdLine.ParseParameters(args)
	if err != nil {
                t.Errorf(err.Error())
        }
	t.Log(config.DnsAddr)
}
