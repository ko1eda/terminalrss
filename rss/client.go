package rss

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Client represents an Rss Client instance
// It can connect to web and xml file sources to load Rss Feeds
// It currently works with Atom Rss Feeds and V2 Rss Feeds.
type Client struct {
	StorageRoot  string
	SortOrder    SortOrder
	Processor    *Processor
	SourceMapper *SourceMapper
	HttpClient   *http.Client
	feed         Feed
}

// New Client configures an Rss Client with sensible defaults.
// Clients can process rss from web sources as well as xml files on your local machine.
// You can set the Storage Root for the client to adjust where these are read from.
func NewClient() (*Client, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err // TODO: Log to file
	}
	c := &http.Client{}
	s, _ := NewSourceMapper()
	storageRoot := filepath.Join(homedir, "terminalrss", "xml")
	return &Client{
		Processor:    &Processor{},
		HttpClient:   c,
		SourceMapper: s,
		feed:         Feed(make([]*Item, 0, 100)),
		StorageRoot:  storageRoot,
		SortOrder:    DATE_DSC,
	}, nil
}

// This updates the clients StorageRoot.
// File urls' should be relative to this path.
func (c *Client) AddStorageRoot(path string) {
	c.StorageRoot = path
}

// Add sources to client to load Rss data from
func (c *Client) AddSources(sources []*Source) {
	c.SourceMapper.AddSources(sources)
}

// Removes any number of sources from the client
func (c *Client) RemoveSources(sources []*Source) {
	c.SourceMapper.RemoveSources(sources)
}

// Returns a slice of all Rss sources the Client contains
func (c *Client) ListSources() []*Source {
	return c.SourceMapper.ListSources()
}

// Find a given source by path or title
// Returns true if we have hit and miss with an empty source if we don't have a match
func (c *Client) FindSource(target string) (*Source, bool) {
	return c.SourceMapper.FindSource(target)
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
		switch source.Type {
		case HTTP:
			go func(source *Source) {
				resp, err := c.HttpClient.Get(source.Path)
				if err != nil {
					log.Fatal(err)
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
		case FILE:
			go func(source *Source) {
				file, err := os.Open(filepath.Join(c.StorageRoot, source.Path))
				if err != nil {
					log.Fatal(err) // add threadsafe logger
				}
				bytes, err := io.ReadAll(file)
				if err != nil {
					log.Fatal(err)
				}
				if err := file.Close(); err != nil {
					log.Fatal(err)
				}
				select {
				case byteChan <- bytes:
					sourceChan <- source
				case <-ctx.Done():
					byteChan <- []byte("")
				}
			}(source)
		}
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
