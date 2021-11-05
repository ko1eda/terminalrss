package rss

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Client struct {
	*http.Client
	Processor *Processor
	RssItems  []*Item
	Sources   []string
}

// TODO add config
// TODO Make sources a map of string to Items
func NewClient() (*Client, error) {
	c := &http.Client{}
	return &Client{
		Client:    c,
		Processor: &Processor{},
		RssItems:  make([]*Item, 0, 500),
		Sources:   make([]string, 0, 100),
	}, nil
}

func (c *Client) AddSources(Sources []string) {
	c.Sources = append(c.Sources, Sources...)
}

// Refresh all rss feeds
// TODO: settings for multithreading vs single
// TODO: make sure multithreading is actually better
// TODO: Remove goroutine from inside anyonmous func, make name function and call that as a goroutine for best practice
func (c *Client) Refresh() {
	loopLen := len(c.Sources)
	byteChan := make(chan []byte, loopLen)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// TODO: look up defer
	for _, url := range c.Sources {
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
		c.RssItems = append(c.RssItems, collection.All()...)
	}

	close(byteChan)

	for i, item := range c.RssItems {
		fmt.Println(item.Date.String() + " " + item.Link)
		fmt.Println(i)
	}
}

func (c *Client) xmlToRss(bytes []byte) (*Collection, error) {
	collection, _ := c.Processor.XmlToRssCollection(bytes)
	return collection, nil
}
