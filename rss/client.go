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
	sourceSlice []*Source
	Sources     map[string]*Source
}

// TODO: add config
func NewClient() (*Client, error) {
	c := &http.Client{}
	return &Client{
		Client:      c,
		Processor:   &Processor{},
		feed:        Feed(make([]*Item, 0, 100)),
		sourceSlice: make([]*Source, 0, 100),
		Sources:     make(map[string]*Source, 100),
		HasSources:  false,
		SortOrder:   DATE_DSC,
	}, nil
}

// Add sources to client to load Rss data from
func (c *Client) AddSources(sources []*Source) {
	m := c.Sources

	if !c.HasSources {
		m = make(map[string]*Source, 100)
		c.HasSources = true
	}

	for _, source := range sources {
		_, hit := m[source.Path]

		if hit {
			continue
		}
		// If the source isn't in our map of sources
		// creat a new rss item slice that we will
		// use when we load rss items
		c.sourceSlice = append(c.sourceSlice, source)
		m[source.Path] = source
	}

	c.Sources = m
}

// Removes any number of sources from the client
func (c *Client) RemoveSources(sources []*Source) {
	for _, source := range sources {
		_, hit := c.Sources[source.Path]

		if !hit {
			continue
		}

		delete(c.Sources, source.Path)
	}

	// If someone passed in a nil slice or empty slice
	// there is no need to recreate the source slice
	if len(sources) < 1 {
		return
	}
	// Here we replace source slices with whatever sources are left in the map
	c.sourceSlice = make([]*Source, 0, len(c.Sources))

	for _, source := range c.Sources {
		c.sourceSlice = append(c.sourceSlice, source)
	}
}

// Returns a slice of all Rss sources the Client contains
func (c *Client) ListSources() []*Source {
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
func (c *Client) Load(sources []*Source) Feed {
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
func (c *Client) load(sources []*Source) Feed {
	if len(sources) < 1 {
		sources = c.ListSources()
	}

	loopLen := len(sources)
	byteChan := make(chan []byte, loopLen)
	sourceChan := make(chan *Source, loopLen)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// TODO: log error to threadsafe logger
	for _, source := range sources {
		go func(source *Source) {
			resp, err := c.Get(source.Path)
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
				sourceChan <- source
			case <-ctx.Done():
				// Send blank bytes into our channel so we don't read forever
				byteChan <- []byte("")
			}
		}(source)
	}

	for i := loopLen; i > 0; i-- {
		bytes := <-byteChan

		if string(bytes) == "" {
			continue
		}

		feed := c.xmlToRssFeed(bytes)
		feed.AddSource(<-sourceChan)
		c.feed = append(c.feed, feed...)
	}

	close(byteChan)
	close(sourceChan)

	return c.SortFeed(c.SortOrder)
}

// Calls the Rss Processors conversion method under the hood
func (c *Client) xmlToRssFeed(bytes []byte) Feed {
	feed, _ := c.Processor.XmlToRssFeed(bytes)
	return feed
}
