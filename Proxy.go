package speed

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type IPList struct {
	IP   string
	Port string
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

func oneToMultiIP(link *url.URL, proxy string, timeout time.Duration, ch chan *http.Response, wg *sync.WaitGroup) struct{} {

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
		return struct{}{}
	}

	defer res.Body.Close()

	ch <- res

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
	ch := make(chan *http.Response) // output
	in := make(chan string)         // input

	for i := 0; i < nWorkers; i++ {
		go worker(in, ch, p, t, &wg)
	}

	for _, x := range ips {
		wg.Add(1)
		in <- x
	}

	go func() {
		wg.Wait()
		// time.Sleep(time.Minute * 3)
		fmt.Println("did i close")
		close(ch)
	}()

	fmt.Println("Have i closed")
}

func worker(jobs <-chan string, results chan *http.Response, link *url.URL, timeout time.Duration, wg *sync.WaitGroup) {
	for x := range jobs {
		oneToMultiIP(link, x, timeout, results, wg)
	}
}
