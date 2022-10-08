package tests

import (
	"testing"
	"time"

	"github.com/taahakh/crunch"
)

type proxyTest struct {
	proxy string
	err   bool
}

// Unit testing the proxies.
// NOTE: Some of the proxy values will not work. Must be updated
var proxies = []proxyTest{
	{"...", false},
	{"...", true},
	{"...", false},
}

func TestProxy(t *testing.T) {
	tyme, err := time.ParseDuration("2s")

	if err != nil {
		t.Errorf(err.Error())
	}

	for _, x := range proxies {
		if _, err = crunch.Proxy("https://httpbin.org/", x.proxy, tyme); x.err == (err == nil) {
			t.Errorf("Proxy error: %s [test -> %t]", err.Error(), x.err)
		}
	}

}
