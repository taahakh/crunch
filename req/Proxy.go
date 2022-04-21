package req

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	// "net/proxy"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

const (
	UntilComplete = ""
)

type IP interface {
	List() []string
}

func GenodeRead(csv [][]string, protocol string) []string {
	var ipList []string

	for i := range csv {
		if i == 0 {
			continue
		}
		ipList = append(ipList, protocol+"://"+csv[i][0]+":"+csv[i][7])
	}

	return ipList
}

func ConnProxNoDefer(link, proxy string, timeout time.Duration) (*http.Response, error) {
	p, err := url.Parse(proxy)
	if err != nil {
		log.Println("Proxy parsing not working")
		return &http.Response{}, err
	}

	l, err := url.Parse(link)
	if err != nil {
		log.Println("Link parsing not working")
		return &http.Response{}, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(p),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	req, err := http.NewRequest("GET", l.String(), nil)
	if err != nil {
		log.Println("New request failed")
		return &http.Response{}, err
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println("Client do not working")
		return &http.Response{}, err
	}

	return res, err
}

func CreateLinkRequestContext(links []*url.URL, retries int) []*RequestItem {
	r := make([]*RequestItem, 0, len(links))

	for _, x := range links {
		ctx, cancel := context.WithCancel(context.Background())
		req, err := http.NewRequestWithContext(ctx, "GET", x.String(), nil)
		if err != nil {
			log.Println("CreateLinkRequestContext: Failed")
		}

		r = append(r, &RequestItem{
			Request: req,
			Cancel:  &cancel,
		})
	}

	return r
}

func CreateProxyClient(proxies []*url.URL, timeout time.Duration) []*http.Client {
	clients := make([]*http.Client, 0, len(proxies))

	for i := 0; i < len(proxies); i++ {
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxies[i]),
		}
		client := &http.Client{
			Transport: transport,
			Timeout:   timeout,
		}
		clients = append(clients, client)
	}
	return clients
}

func ConvertToURL(c []string) []*url.URL {
	urls := make([]*url.URL, 0, len(c))
	for _, x := range c {
		l, err := url.Parse(x)
		if err != nil {
			log.Println("ConvertToURL: Couldn't parse url")
		}
		urls = append(urls, l)
	}
	return urls
}

func CreateNewRequest(method string, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Println(err)
	}

	return req
}

func SimpleContextSetup(proxy []string, urls []string, retries int, timeout time.Duration) *RequestCollection {
	req := ConvertToURL(urls)
	cli := ConvertToURL(proxy)

	ri := CreateLinkRequestContext(req, retries)
	c := CreateProxyClient(cli, timeout)

	rj := &RequestJar{
		Clients: c,
		Links:   ri,
	}

	rs, err := CreateRequestSend(c, ri)
	if err != nil {
		panic("NICE")
	}

	return &RequestCollection{
		RJ: rj,
		RS: rs,
	}
}

func CreateRequestSend(rc []*http.Client, ri []*RequestItem) ([]*RequestSend, error) {
	counter := 0
	if len(rc) == 0 || len(ri) == 0 {
		return nil, errors.New("Either list is of size 0")
	}
	rs := make([]*RequestSend, 0, len(ri))
	for _, x := range ri {
		if counter == len(rc) {
			counter = 0
		}
		rs = append(rs, &RequestSend{
			Request: x,
			Client:  rc[counter],
		})
	}

	return rs, nil
}

func CreateSOCKS5Client(ip string) *http.Client {
	dials, err := proxy.SOCKS5("tcp", ip, nil, proxy.Direct)
	if err != nil {
		fmt.Println("error connecting to proxy", err)
	}
	transport := &http.Transport{
		Dial: dials.Dial,
	}

	client := &http.Client{
		Transport: transport,
	}

	return client
}
