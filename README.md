# Crunch

A Web Scraper that works on mass grouped requests. 

Crunch provides flexibility in creating your own custom retry and request handlers while providing a default solution

Crunch pools provides functionality to control collections externally e.g. end all requests

## Examples

### Without Proxy

```go

func onHTML(x req.Result) bool {
    doc := traverse.HTMLNodeToDoc(x.Document())
    
    // Do whatever
    fmt.Println(doc)
    return true
}

func main() {
    duration, _ := time.ParseDuration("2s")
    
    // No Proxy at all
    c := crunch.NoProxy(
        []string{"..."},
        duration,
        onHTML
    )

    crunch.Do(c, "Run", nil)
}


```

### With Proxy

```go

func onHTML(x req.Result) bool {
    doc := traverse.HTMLNodeToDoc(x.Document())
    
    // Do whatever
    fmt.Println(doc)
    return true
}

func main() {
    duration, _ := time.ParseDuration("2s")
    
    // With proxy/not
    c := crunch.ProxySetup(
        []string{"..."}, []string{"..."}, // Proxy / urls
        nil, 5,                        // Headers / Number of retries
        duration,
        onHTML
    )

    crunch.Do(c, "Run", nil)
}


```
### With Pool

```go
func main() {
    
    c := crunch.ProxySetup(...)

    pool := req.Pool{}
    pool.New("pool", req.PoolSettings{
        AllCollectionsCompleted: func(p PoolLook) {
            ...
        },
    })
    pool.Add("new", c)
    pool.RunSession("new")
}


```

## To-Do

This is an on-going project. There will be bugs but overall, crunch carries out its default functionality.

There are missing features and the current code base will be heavily reworked.

### Request

- Create Pool Manager
- SOCKS compatibility
- Cookie implementation
- Queue system (With chan?)
- Replace http.Clients with http.Transport
- Batch requests
- Merge request module with crunch?
- And more...

### Traverse

- Create proper parser
- Optimise Search functions & Create default search function
- Rework HTMLDocument and its respective functions
- And more...

Benchmarks and Tests need to be provided

crunch v0.1.0
