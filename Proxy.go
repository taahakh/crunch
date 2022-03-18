package speed

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

func ConnWithProx(link, proxy string) {
	p, err := url.Parse(proxy)
	if err != nil {
		log.Println("Proxy parsing not working")
	}

	l, err := url.Parse(link)
	if err != nil {
		log.Println("Link parsing not working")
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(p),
	}

	client := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", l.String(), nil)
	if err != nil {
		log.Println("New request failed")
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println("Client do not working")
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Could not read res.Body")
	}

	log.Println(string(data))
}
