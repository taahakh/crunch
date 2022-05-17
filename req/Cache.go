package req

import (
	"encoding/gob"
	"log"
	"os"

	"github.com/taahakh/speed/traverse"
)

type Cache struct {
	sites    map[string]*traverse.HTMLDocument
	refresh  map[string]int
	nRefresh int
}

func (c *Cache) New() {
	c.sites = make(map[string]*traverse.HTMLDocument)
	c.refresh = make(map[string]int)
}

func (c *Cache) Add(req string, doc *traverse.HTMLDocument) bool {
	if _, ok := c.sites[req]; !ok {
		c.sites[req] = doc
		return true
	}

	return false
}

func (c *Cache) Remove(req string) bool {
	if _, ok := c.sites[req]; ok {
		delete(c.sites, req)
		return true
	}

	return false
}

func (c *Cache) Exists(req string) bool {
	_, ok := c.sites[req]
	return ok
}

func (c *Cache) Save() error {
	wErr := write("./cache.gob", *c)
	if wErr != nil {
		log.Println(wErr)
		return wErr
	}
	return nil
}

func (c Cache) Load() (*Cache, error) {
	rErr := read("./cache.gob", c)
	if rErr != nil {
		log.Println(rErr)
		return nil, rErr
	}
	return &c, rErr
}

func write(path string, obj Cache) error {
	file, err := os.Create(path)
	defer file.Close()
	if err == nil {
		enc := gob.NewEncoder(file)
		enc.Encode(obj)
	}
	return err
}

func read(path string, obj Cache) error {
	file, err := os.Open(path)
	defer file.Close()
	if err == nil {
		dec := gob.NewDecoder(file)
		err = dec.Decode(&obj)
	}
	return err
}
