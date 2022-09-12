package main

import (
	"log"
	"net/http"
	"time"

	"github.com/taahakh/crunch"
	"github.com/taahakh/crunch/req"
	"github.com/taahakh/crunch/traverse"
)

type SaveData []string

// There are three seperate packages
// crunch - Standard methods for scraping and requesting
// req - Methods for requesting webpages. Pools to manage multiple sets of requests
// traverse - Depth first traversal of HTML Nodes
// NOTE: This traverse module is custom made

// traverse package can be replaced by a different package if wanted
// crunch & req packages are needed to carry out basic functionality

// this method tells crunch methods on how to traverse the document and what to do with it
func toScrape(result req.Result) bool {

	// Here we are converting our result document (a *html.Node) into a different structure that
	// we can manipulate
	//
	// the return value of result.Document() is where you can use other packages to carry out DOM
	// traversal
	doc := traverse.HTMLNodeToDoc(result.Document())

	// Print out the text in the HTML File
	log.Println("Whole webpage: ", doc.Node.Text())

	// Examples of what you can do

	// We are finding the node(s) that have those attributes and we have a method that converts those nodes
	// so that we can extract the text within these tags
	//
	// There are a variety of options to choose from. There are different types of 'Find' Methods
	// You can look or sibling children etc., render HTML from notes etc.
	log.Println(doc.Find("[id='fiveDayText']", func(inner *traverse.HTMLDocument) {

		// When developing, you can use this function to see all available nodes that the
		// search function has found
		// inner.PrintNodeList()

		// I want to save my results so im going to create a struct that will contain strings
		// It better to use structs due to the nature of interfaces as we going to have to
		// depolymorphise (don't think that's the right term for this situation) in order to retrieve data
		// It shouldn't be used in practice. However in areas where different structs and different data types are needed
		// to be stored conveniently, this is perfect
		var data SaveData

		for _, x := range inner.Nodify() {

			txt := x.Text()

			log.Println("Specific part of the webpage: ", txt)

			data = append(data, txt)
			result.Save(data)
		}
	}))

	// you can save (store) any struct.
	// this is if you are scraping multiple sites and want a place to store
	// result.Save(nil) -> nil should be replaced with a struct (object) in order to save something

	// if you want to carry out requests on the fly, this method allows you to do so
	// result.Scrape(nil) -> should be replaced by traversing method and urls

	// returns *html.Node. Standard parsed html node
	// result.Document()
	return true
}

func NoProxyMethod() (*req.Collection, error) {
	time, err := time.ParseDuration("5s")

	if err != nil {
		log.Fatal("Failed to parse Duration", err)
	}

	// This is a default method. This method creates a 'collection' which stores all urls,
	// proxies, user-agents, handlers etc.
	// You can create custom collections but certain rules need to be followed in order to work
	//
	// This takes in a list of urls, timeout duration, method on how to traverse document
	col := crunch.NoProxy(
		[]string{"https://metoffice.gov.uk"},
		time,
		toScrape,
	)

	// Here you can run your collections
	// col -> collection
	// Run -> methodology of scrape/requests
	// nil -> A SessionHandler. You can customise how you want IP, User-Agent rotations to occur and
	// 		  other requirements when needed. A default handler is put in place. IT IS NOT SOPHISTICATED AND MUST BE REPLACED.
	return crunch.Do(col, "Run", nil)
}

// Note that this method will not work
// Appropriate values are needed. This is just a demonstration.
func WithOrWithoutProxy() {
	duration, _ := time.ParseDuration("2s")

	// With proxy/not
	c := crunch.ProxySetup(
		[]string{"..."}, []string{"..."}, // Proxy / urls
		nil, 5, // Headers / Number of retries
		duration,
		toScrape,
	)

	// You can append clients (Proxies) to the collection.
	// Also headers etc.
	// c.RJ.Clients = append(c.RJ.Clients, &http.Client{})

	// I have applied DefaultSessionHandler. There is no needed as it will be automatically applied
	// when no other handler has been added
	crunch.Do(c, "Run", &req.DefaultSessionHandler{})
}

// Pools of collections. This is useful for automating scrape sessions when needed.
// Servers can use pools to manage multiple collections and carry out scrapes whenever needed
// NOTE: When the pool is used, the pool itself is not running concurrently. It is running sequentially.
// However, when running its collections, the COLLECTIONS are run concurrenly.
// Some methods with the Pool are 'Blocking' so any sequential code that you want to reach will be executed later once the blocking call is finished
// It is best to run the Pool itself in a goroutine and use the main method to control everything
func TryWithPool() {
	duration, _ := time.ParseDuration("2s")

	// Note this will not work
	c := crunch.ProxySetup(
		[]string{""},
		[]string{""},
		[]*http.Header{},
		0, duration, nil,
	)

	pool := req.Pool{}

	// We name the pool and we can add settings to our pool
	// There are default implementations in place
	pool.New("pool", req.PoolSettings{
		AllCollectionsCompleted:      func(p req.PoolLook) {},
		IncomingCompletedCollections: func(rc *req.Collection) {},
	})

	// Add collections to the pool
	pool.Add("new", c)

	// Run the collection. You can run the collection with different methods
	pool.RunSession("new", nil)

	// The pool can manage all the collections. It can add new collections, cancell all or specific
	// collections. See what data has been scraped. etc.
	// pool.AmIDone("new")
	// pool.CancelCollection("new")

}

func main() {
	// This returns the collection.
	// We can see our stored results here
	data, _ := NoProxyMethod()
	log.Println(data.Result)

	// Note this will not work without the proper values
	// Due to the fact that we need a list of proxy ips that will probably not work soon
	// WithOrWithoutProxy()

	// Note this will not work without the proper values
	// Due to the fact that we need a list of proxy ips that will probably not work soon
	// TryWithPool()
}
