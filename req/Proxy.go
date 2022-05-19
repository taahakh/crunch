package req

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

const (
	UntilComplete = ""
)

/* READERS */
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

func SingleList(csv [][]string, protocol string) []string {
	var ipList []string

	for i := range csv {
		if i == 0 {
			continue
		}
		ipList = append(ipList, protocol+"://"+csv[i][0])
	}

	return ipList
}

/* CONNECTION */

// Proxy is a simple request made with a proxy
//
// ONLY HTTP/HTTPS compatable
func Proxy(link, proxy string, timeout time.Duration) (*http.Response, error) {
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

/* STRUCTURES */

// MakeRequestItem takes in urls and creates a struct that contains http.Request and cancellation function
func MakeRequestItems(links []*url.URL) []*RequestItem {
	r := make([]*RequestItem, 0, len(links))

	for _, x := range links {
		ctx, cancel := context.WithCancel(context.Background())
		req, err := http.NewRequestWithContext(ctx, "GET", x.String(), nil)
		if err != nil {
			log.Println("RequestItem: Failed")
		}

		r = append(r, &RequestItem{
			Request: req,
			Cancel:  &cancel,
		})
	}

	return r
}

// MakeProxyClient creates a list of clients with attached timeouts
func MakeProxyClients(proxies []*url.URL, timeout time.Duration) []*http.Client {
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

// ConvertToURL converts strings to appropriate urls
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

// MakeRequestItem creates a http.Request with cancellation function
func MakeRequestItem(link string) *RequestItem {
	x, err := url.Parse(link)

	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, "GET", x.String(), nil)

	if err != nil {
		log.Println("CreateLinkRequestContext: Failed")
	}

	return &RequestItem{
		Request: req,
		Cancel:  &cancel,
	}
}

// ItemToSend puts RequestItems in Send struct. Scrape functionality is added with these structs
func ItemToSend(items []*RequestItem, m func(rp Result) bool) []*Send {
	send := make([]*Send, 0, len(items))
	for _, x := range items {
		send = append(send, &Send{
			Request: x,
			Retries: 5,
			Scrape:  m,
		})
	}
	return send
}

// NoProxy doesn't use proxy to make requests
//
// Uses shared client for all requests
func NoProxy(urls []string, timeout time.Duration, method func(rp Result) bool) *Collection {
	rs := make([]*Send, 0, len(urls))
	req := ConvertToURL(urls)

	client := &http.Client{}

	ri := MakeRequestItems(req)
	for _, x := range ri {

		rs = append(rs, &Send{
			Request: x,
			Scrape:  method,
			Client:  client,
		})
	}

	return &Collection{
		RS: rs,
		RJ: &Jar{
			Clients: []*http.Client{{}},
		},
	}
}

// ProxySetup allows requests to be dialled through a proxy
//
// Proxy, Url String lists and timeout must not be empty
// Method func and retries should be added to make the request and scrape more useful
func ProxySetup(
	proxy []string,
	urls []string,
	headers []*http.Header,
	retries int,
	timeout time.Duration,
	method func(rp Result) bool) *Collection {

	var ri []*RequestItem

	if proxy == nil || urls == nil || len(proxy) == 0 || len(urls) == 0 {
		return nil
	}

	req := ConvertToURL(urls)
	cli := ConvertToURL(proxy)

	ri = MakeRequestItems(req)
	c := MakeProxyClients(cli, timeout)

	if headers != nil {
		req, err := ApplyHeadersRI(ri, headers)
		if err != nil {
			log.Println("Headers could not be applied")
		}
		ri = req
	}

	rj := &Jar{
		Clients: c,
		Headers: headers,
	}

	rs, err := MakeSends(c, ri, retries, method)
	if err != nil {
		panic("NICE")
	}

	return &Collection{
		RJ: rj,
		RS: rs,
	}
}

// MakeSends attaches client, requests, retries, scraping method in one struct
func MakeSends(rc []*http.Client, ri []*RequestItem, retries int, method func(rp Result) bool) ([]*Send, error) {
	counter := 0
	if len(rc) == 0 || len(ri) == 0 {
		return nil, errors.New("Either list is of size 0")
	}
	rs := make([]*Send, 0, len(ri))
	for _, x := range ri {
		if counter == len(rc) {
			counter = 0
		}
		rs = append(rs, &Send{
			Request: x,
			Client:  rc[counter],
			Scrape:  method,
			Retries: retries,
		})
	}

	return rs, nil
}

// SOCKS5Client returns http.Client for use of sock proxy
func SOCKS5Client(ip string) *http.Client {
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

// ApplyHeadersRI links headers to requests
func ApplyHeadersRI(req []*RequestItem, headers []*http.Header) ([]*RequestItem, error) {
	counter := 0
	length := len(headers)
	if length == 0 {
		return req, errors.New("Headers arr is empty")
	}
	for _, x := range req {
		if counter == length {
			counter = 0
		}
		x.Request.Header = *headers[counter]
	}

	return req, nil
}

// ApplyHeaders applies headers to requests
func ApplyHeaders(req []*http.Request, headers []*http.Header) ([]*http.Request, error) {
	counter := 0
	length := len(headers)
	if length == 0 {
		return req, errors.New("Headers arr is empty")
	}
	for _, x := range req {
		if counter == length {
			counter = 0
		}
		x.Header = *headers[counter]
	}

	return req, nil
}

// ApplyUserAgents applies User-Agent header to header list
func ApplyUserAgents(headers []*http.Header, agents []string) ([]*http.Header, error) {
	counter := 0
	length := len(agents)
	if length == 0 {
		return nil, errors.New("There are no agents")
	}
	for _, x := range headers {
		if counter == length {
			counter = 0
		}
		x.Set("User-Agent", agents[counter])
	}

	return headers, nil
}

// CreateHeaders transfroms list into http.Header
//
// This doesn't need to be used if in the format of map[string][]string
func CreateHeaders(agents []string) ([]*http.Header, error) {
	headers := make([]*http.Header, 0, len(agents))

	if len(agents) == 0 {
		return nil, errors.New("Empty agents")
	}

	for _, x := range agents {
		h := CreateHeader(x)
		headers = append(headers, &h)
	}

	return headers, nil
}

// CreateHeader creates a simple header struct that takes a User-Agent as its first input
func CreateHeader(agent string) http.Header {
	return http.Header{
		"User-Agent": []string{agent},
	}
}
