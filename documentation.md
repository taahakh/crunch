# Request

## Main structs

### Collection

Stores all requests, clients, headers etc. Used by request handlers to start making request and start scraping.


### Pool

An external handler that can control collections and requests through goroutines and channels. A good way to automate scraping/requests etc. It 'pools' similar collections together that require similar functionality.


## Request handlers

### Run

Handles requests from collection stored in RS. There are no retry handling for detected bots / rotating ip

### Session

Same as Run but has handlers for retry. Collection stores all Proxies and Headers which you can independently store in conjuction with the handler. 

You can use the handler to connect to another resource to handle retries if you don't want to store the data in collection struct.

However, all requests must be stored in the collection



# Traverse

## Search

tag[attr='val']

### Find


### FindStrictly


### Query