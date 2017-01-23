package utils

import (
	"testing"
)

func TestInitLoggers(t *testing.T) {
	var err error
	err = InitLoggers(0)
	if err != nil {
		t.Error("Unable to initialize loggers! %s", err.Error())
	}
}

func TestInitLoggers1(t *testing.T) {
	var err error
	err = InitLoggers(1)
	if err != nil {
		t.Error("Unable to initialize loggers! %s", err.Error())
	}
}

func TestInitLoggers2(t *testing.T) {
	var err error
	err = InitLoggers(5)
	if err != nil {
		t.Error("Unable to initialize loggers! %s", err.Error())
	}
}
