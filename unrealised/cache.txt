package req

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/taahakh/speed/traverse"
)

// Set time/day on when needed to refresh
type Cache struct {
	Sites    map[string]traverse.HTMLDocument
	Refresh  map[string]int
	NRefresh int
}

func (c Cache) New() *Cache {
	c.Sites = make(map[string]traverse.HTMLDocument)
	c.Refresh = make(map[string]int)
	c.NRefresh = 0
	return &c
}

func (c *Cache) Add(req string, doc traverse.HTMLDocument) bool {
	if _, ok := c.Sites[req]; !ok {
		c.Sites[req] = doc
		return true
	}
	return false
}

// ./cache.gob
func (c *Cache) Save(path string) error {
	wErr := write(path, *c)
	if wErr != nil {
		log.Println("Saving: ", wErr)
		return wErr
	}
	return nil
}

func (c *Cache) Has(req string) (*traverse.HTMLDocument, error) {
	if val, ok := c.Sites[req]; ok {
		return &val, nil
	}
	return nil, errors.New("Doesn't contain document")
}

func (c Cache) Load(path string) (*Cache, error) {
	rErr := read(path, &c)
	if rErr != nil {
		log.Println("Loading: ", rErr)
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

func read(path string, c *Cache) error {
	file, err := os.Open(path)
	defer file.Close()
	if err == nil {
		dec := gob.NewDecoder(file)
		err = dec.Decode(c)
		fmt.Println("did i looad: ", c)
	}
	return err
}
