package rss

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"
)

type Client struct {
	*http.Client
	HasSources  bool
	Processor   *Processor
	feed        Feed
	SortOrder   SortOrder
	sourceSlice []string
	Sources     map[string][]*Item
}

// TODO: add config
func NewClient() (*Client, error) {
	c := &http.Client{}
	return &Client{
		Client:      c,
		Processor:   &Processor{},
		feed:        Feed(make([]*Item, 0, 100)),
		sourceSlice: make([]string, 0, 100),
		Sources:     make(map[string][]*Item, 100),
		HasSources:  false,
		SortOrder:   DATE_DSC,
	}, nil
}

// Add sources to client to load Rss data from
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
		c.sourceSlice = append(c.sourceSlice, source)
		m[source] = make([]*Item, 0, 100)
	}

	c.Sources = m
}

// Removes any number of sources from the client
func (c *Client) RemoveSources(sources []string) {
	for _, source := range sources {
		_, hit := c.Sources[source]

		if !hit {
			continue
		}

		delete(c.Sources, source)
	}

	// If someone passed in a nil slice or empty slice
	// there is no need to recreate the source slice
	if len(sources) < 1 {
		return
	}
	// Here we replace source slices with whatever sources are left in the map
	c.sourceSlice = make([]string, 0, len(c.Sources))

	for source := range c.Sources {
		c.sourceSlice = append(c.sourceSlice, source)
	}
}

// Returns a slice of all Rss sources the Client contains
func (c *Client) ListSources() []string {
	return c.sourceSlice
}

// Returns the internal Feed type.
// This Should be used after at least one call to Refresh, Load or LoadAll
func (c *Client) Feed() Feed {
	return c.feed
}

// Sort internal Feed by the given SortOrder and return the feed
func (c *Client) SortFeed(order SortOrder) Feed {
	c.SortOrder = order
	c.feed.SortBy(order)
	return c.Feed()
}

// Load RSS Feed Items into the clients Feed from a given source or sources.
// If no sources given this loads all sources.
// We can use this as a subslice from ListSources to load only from specific sources
func (c *Client) Load(sources []string) Feed {
	if len(sources) < 1 {
		return c.LoadAll()
	}
	return c.load(sources)
}

// Refresh all rss feeds
func (c *Client) Refresh() Feed {
	return c.LoadAll()
}

// Load Rss Feed Items into the clients Feed from all sources
func (c *Client) LoadAll() Feed {
	return c.load(nil)
}

// Internal method used to load sources asynchronously from a number of sources
func (c *Client) load(sources []string) Feed {
	if len(sources) < 1 {
		sources = c.ListSources()
	}

	loopLen := len(sources)
	byteChan := make(chan []byte, loopLen)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// TODO: log error to threadsafe logger
	for _, url := range sources {
		go func(url string) {
			resp, err := c.Get(url)
			if err != nil {
				log.Fatal(err)
				// cancel() // we would only want to call this here if we immediately want to cancel all calls (we would only do this on fatal error otherwise logging is fine)
			}
			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			select {
			case byteChan <- bytes:
			case <-ctx.Done():
				// Send blank bytes into our channel so we don't read forever from our channel
				byteChan <- []byte("")
			}
		}(url)
	}

	for i := loopLen; i > 0; i-- {
		bytes := <-byteChan

		if string(bytes) == "" {
			continue
		}

		collection, _ := c.xmlToRss(bytes)
		c.feed = append(c.feed, collection.All()...)
	}
	close(byteChan)

	return c.SortFeed(c.SortOrder)
}

// Calls the Rss Processors conversion method under the hood
func (c *Client) xmlToRss(bytes []byte) (*Collection, error) {
	collection, _ := c.Processor.XmlToRssCollection(bytes)
	return collection, nil
}
