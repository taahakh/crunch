package speed

import (
	"fmt"
	"io"
	"log"
	"net/http"

	// "net/proxy"
	"net/url"
	"runtime"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

const (
	UntilComplete = ""
)

// Groups of ips and links
type RequestJar struct {
	Clients []*http.Client
	Links   []*http.Request // this is intially in the form of url.URL but is then converted to string
}

type RequestSend struct {
	// Retries <= 0  - tries unless finished
	Client  *http.Client
	Request *http.Request
	Retries int
}

type RequestCollection struct {
	// Finish tells us when we want the webscrape to end by no matter what
	// Finish nil will go on until everything is finished
	RJ     *RequestJar
	RS     []*RequestSend
	Finish string // how long it should take before the rc should end
}

type RequestResult struct {
	mu      sync.Mutex
	res     []HTMLDocument
	counter int
}

type GroupRequest struct {
	Ips     []string
	Links   []string
	Timeout time.Duration
}

type SingleRequest struct {
	proxyClient *http.Client
	link        *http.Request
	retries     int
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

func oneToMultiIP(link *url.URL, proxy string, timeout time.Duration, ch chan *http.Response, done chan struct{}, wg *sync.WaitGroup) struct{} {

	defer wg.Done()

	p, err := url.Parse(proxy)
	if err != nil {
		log.Println("Proxy parsing not working")
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(p),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	req, err := http.NewRequest("GET", link.String(), nil)
	if err != nil {
		log.Println("New request failed")
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println("Client do not working")
		done <- struct{}{}
		return struct{}{}
	}

	defer res.Body.Close()

	ch <- res
	done <- struct{}{}

	return struct{}{}
}

func LinktoMultiIP(dest, timeout string, ips []string, nWorkers int) {
	t, err := time.ParseDuration(timeout)
	if err != nil {
		log.Println("no parse")
	}

	p, err := url.Parse(dest)
	if err != nil {
		log.Println("Destination parsing failed")
	}

	var wg sync.WaitGroup
	ch := make(chan *http.Response, 1) // output
	in := make(chan string, 10)        // input
	done := make(chan struct{}, 10)    // finish
	end := make(chan struct{}, 10)     // finish
	var data *http.Response

	for i := 0; i < nWorkers; i++ {
		go worker(in, ch, done, end, p, t, &wg)
	}

	counter := 1
	goFinish := 1
	in <- ips[counter-1]

loop:
	for {
		if goFinish == nWorkers {
			accept := 0
			for {
				select {
				case data = <-ch:
					end <- struct{}{}
					break loop
				case <-done:
					accept++
				}
				if goFinish == accept {
					goFinish = 0
					break
				}
			}
		}

		if counter == len(ips) {
			break
		}
		counter++
		goFinish++
		in <- ips[counter-1]

	}

	go func() {
		wg.Wait()
		fmt.Println("did i close")
		close(ch)
		close(end)
	}()

	fmt.Println(data)

	fmt.Println("Have i closed")
	fmt.Println(runtime.NumGoroutine())
}

func worker(jobs <-chan string, results chan *http.Response, done chan struct{}, end chan struct{}, link *url.URL, timeout time.Duration, wg *sync.WaitGroup) {
	for {
		select {
		case <-end:
			return
		case x := <-jobs:
			wg.Add(1)
			oneToMultiIP(link, x, timeout, results, done, wg)
		}
	}
}

func ProxyConnection(req *SingleRequest, ch chan *SingleRequest, done chan struct{}, rr *RequestResult, wg *sync.WaitGroup) struct{} {
	defer wg.Done()

	client := req.proxyClient
	request := req.link

	resp, err := client.Do(request)
	if err != nil {
		log.Println("ProxyConnection: Client failed")
		req.decrement()
		ch <- req
		return struct{}{}
	}

	fmt.Println("------------------------------PASSSEDDDDDDDDDDD-------------------------------")

	defer resp.Body.Close()

	data, err := HTMLDocUTF8(resp)
	if err != nil {
		log.Println("Couldn't read body")
		return struct{}{}
	}

	rr.add(data)
	done <- struct{}{}

	return struct{}{}
}

func groupWorker(req <-chan *SingleRequest, out chan *SingleRequest, done chan struct{}, rr *RequestResult, wg *sync.WaitGroup) {
	for x := range req {
		wg.Add(1)
		ProxyConnection(x, out, done, rr, wg)
	}
}

func GroupScrape(gr GroupRequest, nWorkers, retries int) *RequestResult {
	proxies := ConvertToURL(gr.Ips)
	links := ConvertToURL(gr.Links)
	pc := CreateProxyClient(proxies, gr.Timeout)
	lr := CreateLinkRequest(links)

	var wg sync.WaitGroup
	var rr RequestResult
	req := make(chan *SingleRequest, 10)
	out := make(chan *SingleRequest, 10)
	done := make(chan struct{}, 5)

	for i := 0; i < nWorkers; i++ {
		go groupWorker(req, out, done, &rr, &wg)
	}

	counter := 0
	for _, x := range lr {
		req <- &SingleRequest{
			proxyClient: pc[counter],
			link:        x,
			retries:     retries,
		}
		counter++
		if counter == len(pc) {
			counter = 0
		}
	}

	doneCounter := 0
loop:
	for {
		select {
		case item := <-out:
			if item.retries == 0 {
				done <- struct{}{}
				continue
			} else {
				if counter == len(pc) {
					counter = 0
				}
				item.proxyClient = pc[counter]
				req <- item
				counter++
			}
		case <-done:
			doneCounter++
			if doneCounter == len(lr) {
				break loop
			}
		}
	}

	go func() {
		wg.Wait()
		close(out)
		close(req)
		fmt.Println("Closing")
		return
	}()

	return &rr
}

func CreateLinkRequest(links []*url.URL) []*http.Request {
	requests := make([]*http.Request, 0, len(links))
	for _, x := range links {
		req, err := http.NewRequest("GET", x.String(), nil)
		if err != nil {
			log.Println("CreateLinkRequest: Cannot create new request")
		}
		requests = append(requests, req)
	}
	return requests
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

func (c *RequestResult) add(b HTMLDocument) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.res = append(c.res, b)
	c.counter++
}

func (c *RequestResult) Read() []HTMLDocument {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.res
}

func (c *RequestResult) Count() int {
	return c.counter
}

func (rt *SingleRequest) decrement() {
	rt.retries -= 1
}

func CreateNewRequest(method string, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Println(err)
	}

	return req
}

func SimpleSetup(proxy []string, urls []string, timeout time.Duration, retries int) *RequestCollection {
	var req []*http.Request // links
	var cli []*http.Client  // proxies

	reqUrl := ConvertToURL(urls)
	cliUrl := ConvertToURL(proxy)

	req = CreateLinkRequest(reqUrl)
	cli = CreateProxyClient(cliUrl, timeout)

	rj := &RequestJar{
		Clients: cli,
		Links:   req,
	}

	rs := rj.CreateHandle(retries)

	return &RequestCollection{
		RJ: rj,
		RS: rs,
	}
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
