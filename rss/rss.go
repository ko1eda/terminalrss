package rss

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const ()

type Client struct {
	*http.Client
	HasSources  bool
	Processor   *Processor
	Items       []*Item
	sourceSlice []string
	Sources     map[string][]*Item
}

// TODO add config
// TODO Make sources a map of string to Items
func NewClient() (*Client, error) {
	c := &http.Client{}
	return &Client{
		Client:    c,
		Processor: &Processor{},
		Items:     make([]*Item, 0, 500),
		// sourceSlice is used internally
		// to main the ordering of sources in ListSources
		sourceSlice: make([]string, 0, 100),
		// Sorter
		Sources:    make(map[string][]*Item, 100),
		HasSources: false,
	}, nil
}

// This can be passed to NewClient function or used as a standalone function to change the sort used by the client
// func WithSort()
// We should add sources to the client before we preform any other method calls or they will return no RSS Items
func (c *Client) AddSources(sources []string) {
	m := c.Sources

	if !c.HasSources {
		m = make(map[string][]*Item, 100)
		c.HasSources = true
	}

	for _, source := range sources {
		_, hit := m[source]

		if hit {
			continue
		}

		// If the source isn't in our map of sources
		// creat a new rss item slice that we will
		// use when we load rss items
		m[source] = make([]*Item, 0, 100)
		c.sourceSlice = append(c.sourceSlice, source)
	}

	c.Sources = m
}

// Returns a slice of all Rss sources the Client contains
// TODO make this a noop closer or something with bytes read in??
func (c *Client) ListSources() []string {
	return c.sourceSlice
}

// Load RSS Feed Items into the clients Feed from A given source or sources. If no sources given this loads all sources
// Can can use this as a subslice from ListSources to load the specific sources from their urls
// TODO: settings for multithreading vs single
// TODO: make sure multithreading is actually better
// TODO: Remove goroutine from inside anonymous func, make name function and call that as a goroutine for best practice
func (c *Client) LoadFrom(sources []string) {
	if len(sources) < 1 {
		c.LoadAll()
		return
	}
	c.load(sources)
}

// Refresh all rss feeds
func (c *Client) Refresh() {
	c.LoadAll()
}

// Load Rss Feed Items into the clients Feed from all sources
func (c *Client) LoadAll() {
	c.load([]string{})
}

// Internal method used to load sources asynchronously from a number of sources
func (c *Client) load(sources []string) {
	if len(sources) < 1 {
		sources = c.ListSources()
	}

	loopLen := len(sources)
	byteChan := make(chan []byte, loopLen)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// TODO: look up defer
	for _, url := range sources {
		go func(url string) {
			resp, err := c.Get(url)
			if err != nil {
				log.Fatal(err) // TODO: log error to threadsafe logger
				// cancel() // we would only want to call this here if we immediately want to cancel all calls (we would only do this on fatal error otherwise logging is fine)
			}
			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err) // TODO: log error to threadsafe logger
			}
			select {
			case byteChan <- bytes:
			case <-ctx.Done():
				byteChan <- []byte("") // send blank bytes into our channel so we dont have a a deadlock
			}
		}(url)
	}

	for i := loopLen; i > 0; i-- {
		bytes := <-byteChan
		// if we have an empty grouping of bytes we just ignore it
		if string(bytes) == "" {
			// println("ERR")
			continue
		}
		collection, _ := c.xmlToRss(bytes)
		c.Items = append(c.Items, collection.All()...)
	}

	close(byteChan)

	for i, item := range c.Items {
		fmt.Println(item.Date.String() + " " + item.Link)
		fmt.Println(i)
	}
}

func (c *Client) xmlToRss(bytes []byte) (*Collection, error) {
	collection, _ := c.Processor.XmlToRssCollection(bytes)
	return collection, nil
}
