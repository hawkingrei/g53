/* http_test.go
 *
 * Copyright (C) 2016 Alexandre ACEBEDO
 *
 * This software may be modified and distributed under the terms
 * of the MIT license.  See the LICENSE file for details.
 */

package servers

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hawkingrei/dtdns/src/utils"
)

func TestServiceRequests(t *testing.T) {
	const TestAddr = "127.0.0.1:9980"

	config := utils.NewConfig()
	config.HttpAddr = TestAddr

	server := NewHTTPServer(config, NewDNSServer(config))
	go server.Start()

	// Allow some time for server to start
	time.Sleep(250 * time.Millisecond)

	var tests = []struct {
		method, url, body, expected string
		status                      int
	}{
		//{"GET", "/services", "", "{}", 200},
		//{"GET", "/services/foo", "", "", 404},
		//{"PUT", "/services/foo", `{"Aliases": "foo"}`, "", 500},
		{"PUT", "/services/foo", `{"Record_type":"A","Value":"127.0.0.1","TTL":3600,"Aliases":"foo.duitang.com."}`, "", 200},
		{"GET", "/services/foo", "", `{"Record_type":"A","Value":"127.0.0.1","TTL":3600,"Aliases":"foo.duitang.com."}`, 200},
		{"PUT", "/services/boo", `{"Record_type":"A","Value":"127.0.0.2","TTL":3600,"Aliases":"boo.duitang.com."}`, "", 200},
		{"GET", "/services", "", `{"boo":{"Record_type":"A","Value":"127.0.0.2","TTL":3600,"Aliases":"boo.duitang.com."},"foo":{"Record_type":"A","Value":"127.0.0.1","TTL":3600,"Aliases":"foo.duitang.com."}}`, 200},
		//{"PATCH", "/services/boo", `{"Aliases": "bar", "ttl": 20}`, "", 200},
		//{"GET", "/services/boo", "", `{"Aliases":"bar","Value":"127.0.0.2","TTL":20}`, 200},
		//{"DELETE", "/services/foo", ``, "", 200},
		//{"GET", "/services", "", `{"boo":{"Aliases":"bar","Value":"127.0.0.2","TTL":20}}`, 200},
	}

	for _, input := range tests {
		t.Log(input.method, input.url)
		req, err := http.NewRequest(input.method, "http://"+TestAddr+input.url, strings.NewReader(input.body))
		if err != nil {
			t.Error(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()

		if input.status != resp.StatusCode {
			t.Error(input, "Expected status:", input.status, "Got:", resp.StatusCode)
		}

		if input.status != 200 {
			continue
		}

		actual, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		actualStr := strings.Trim(string(actual), " \n")
		if actualStr != input.expected {
			t.Error(input, "Expected:", input.expected, "Got:", actualStr)
		}
	}

	t.Log("Test TTL setter")
	if config.Ttl != 0 {
		t.Error("Default TTL is not 0")
	}
	req, err := http.NewRequest("PUT", "http://"+TestAddr+"/set/ttl", strings.NewReader("12"))
	if err != nil {
		t.Error(err)
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	if config.Ttl != 12 {
		t.Error("TTL not updated. Expected: 12 Got:", config.Ttl)
	}
}
