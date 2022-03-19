package speed

import (
	"fmt"
	"io"
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

// func oneToMultiIP(link, proxy string, timeout time.Duration, ch chan *http.Response, rec chan bool, wg *sync.WaitGroup) struct{} {

// 	defer wg.Done()

// 	go func() struct{} {
// 		for {
// 			select {
// 			case <-rec:
// 				close(ch)
// 				return struct{}{}
// 			}

// 		}
// 	}()

// 	p, err := url.Parse(proxy)
// 	if err != nil {
// 		log.Println("Proxy parsing not working")
// 	}

// 	l, err := url.Parse(link)
// 	if err != nil {
// 		log.Println("Link parsing not working")
// 	}

// 	transport := &http.Transport{
// 		Proxy: http.ProxyURL(p),
// 	}

// 	client := &http.Client{
// 		Transport: transport,
// 		Timeout:   timeout,
// 	}

// 	req, err := http.NewRequest("GET", l.String(), nil)
// 	if err != nil {
// 		log.Println("New request failed")
// 	}

// 	res, err := client.Do(req)
// 	if err != nil {
// 		log.Println("Client do not working")
// 		return struct{}{}
// 	}

// 	ch <- res
// 	rec <- true

// 	return struct{}{}
// }

func oneToMultiIP(link, proxy string, timeout time.Duration, ch chan *http.Response, rec chan bool, wg *sync.WaitGroup) struct{} {

	defer wg.Done()

	go func() struct{} {
		for {
			select {
			case <-rec:
				close(ch)
				return struct{}{}
			}

		}
	}()

	res, err := ConnProxNoDefer(link, proxy, timeout)
	if err != nil {
		return struct{}{}
	}

	ch <- res
	rec <- true

	return struct{}{}
}

func LinktoMultiIP(dest string, ips []string, timeout string) {
	t, err := time.ParseDuration(timeout)
	if err != nil {
		log.Println("no parse")
	}

	var wg sync.WaitGroup
	ch := make(chan *http.Response)
	rec := make(chan bool)
	str := make([]*http.Response, 0)

	for _, ip := range ips {
		wg.Add(1)
		go oneToMultiIP(dest, ip, t, ch, rec, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for item := range ch {
		str = append(str, item)
	}

	for i, x := range str {
		defer x.Body.Close()
		d, err := io.ReadAll(x.Body)
		if err != nil {
			log.Println(err)
		}
		fmt.Println(i, string(d))
	}
}
