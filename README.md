# Crunch

A Web Scraper that works on doing mass grouped requests. 

Create your own custom retry and request handlers.

Use pools to store similar collections that require similiar handling

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

    crunch.Do("Run", c, nil)
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
        []string{"..."}, []string{""}, // Proxy / urls
        nil, 5,                        // Headers / Number of retries
        duration,
        onHTML
    )

    crunch.Do("Run", c, nil)
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