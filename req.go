package speed

import (
	"io"
	"log"
	"net/http"
	// "strings"
)

var (
	// dfs func(x *html.Node, data string)

	MinHeaderList = make(map[string]string)
	HeaderList    = make([]Headers, 0)
	DocumentList  = make([]Doc, 0)
	// TagTrackerList = make([]TagTracker, 0)
)

type Headers struct {
	UserAgent               string
	Accept                  string
	AcceptLanguage          string
	AcceptEncoding          string
	Referer                 string
	Connection              string
	UpgradeInsecureRequests string
	IfModifiedSince         string
	IfNoneMatch             string
	CacheControl            string
}

// returns bytes. needs to be converted to string in order to be readable to users
func Get(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Failed to get LINK")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to READ")
	}
	return body
}

// func (r *Doc) AddTag(t Tag) {
// 	r.Tags = append(r.Tags, t)
// }

// returns utf-8 string text
func GetHtml(url string) string {
	return string(Get(url))
}
