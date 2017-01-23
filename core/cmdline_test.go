package core

import (
	"testing"
)

func TestCmdline(t *testing.T) {
	args := []string{`--nameserver="8.8.8.8:53"`}
	var cmdLine CommandLine
	cmdLine.ParseParameters(args)
}
