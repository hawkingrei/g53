package utils

import (
	"reflect"
	"testing"
)

func TestDomainCreation(t *testing.T) {
	var tests = map[string]string{
		"foo":           "foo",
		"foo.":          "foo",
		".foo.docker.":  "foo.docker",
		".foo..docker.": "foo.docker",
		"foo.docker..":  "foo.docker",
	}

	for input, expected := range tests {
		t.Log(input)
		d := NewDomain(input)
		if actual := d.String(); actual != expected {
			t.Error(input, "Expected:", expected, "Got:", actual)
		}
	}
	input := "127.0.0.1,127.0.0.2"
	expected := nameservers{input}
	if actual := []string{"127.0.0.1", "127.0.0.2"}; reflect.DeepEqual(actual, expected) {
		t.Error(input, "Expected:", expected, "Got:", actual)
	}
}

func TestNewConfig(t *testing.T) {
	config := NewConfig()
	if !reflect.DeepEqual(config.DnsAddr, ":53") {
		t.Error("DnsAddr error")
	}
}
