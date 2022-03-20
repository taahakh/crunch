package speed

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"time"
)

type IPList struct {
	IP   string
	Port string
}

type ReqResult struct {
	mu      sync.Mutex
	res     [][]byte
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

func ProxyConnection(req *SingleRequest, ch chan *SingleRequest, done chan struct{}, rr *ReqResult, wg *sync.WaitGroup) struct{} {
	defer wg.Done()

	client := req.proxyClient
	request := req.link

	// fmt.Println("Client: ", client.Transport)
	// fmt.Println("Request: ", request)

	resp, err := client.Do(request)
	if err != nil {
		log.Println("ProxyConnection: Client failed")
		req.decrement()
		ch <- req
		// wg.Done()
		return struct{}{}
	}

	fmt.Println("------------------------------PASSSEDDDDDDDDDDD-------------------------------")

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Couldn't read body")
	}

	rr.add(data)
	done <- struct{}{}

	return struct{}{}
}

func groupWorker(req <-chan *SingleRequest, out chan *SingleRequest, done chan struct{}, rr *ReqResult, wg *sync.WaitGroup) {
	for x := range req {
		wg.Add(1)
		ProxyConnection(x, out, done, rr, wg)
	}
}

func GroupScrape(gr GroupRequest, nWorkers, retries int) *ReqResult {
	proxies := ConvertToURL(gr.Ips)
	links := ConvertToURL(gr.Links)
	pc := CreateProxyClient(proxies, gr.Timeout)
	lr := CreateLinkRequest(links)

	var wg sync.WaitGroup
	var rr ReqResult
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

	for i := 0; i < len(rr.res); i++ {
		fmt.Println(i, "ok")
	}

	return &rr

}

func CreateLinkRequest(links []*url.URL) []*http.Request {
	var requests []*http.Request
	// requests := make([]*http.Request, len(links))
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
	var clients []*http.Client
	// clients := make([]*http.Client, len(proxies))

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
	var urls []*url.URL
	// urls := make([]*url.URL, len(c))
	for _, x := range c {
		l, err := url.Parse(x)
		if err != nil {
			log.Println("ConvertToURL: Couldn't parse url")
		}
		urls = append(urls, l)
	}
	return urls
}

func (c *ReqResult) add(b []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.res = append(c.res, b)
	c.counter++
}

func (c *ReqResult) Read() [][]byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.res
}

func (c *ReqResult) Count() int {
	return c.counter
}

func (rt *SingleRequest) decrement() {
	rt.retries -= 1
}
