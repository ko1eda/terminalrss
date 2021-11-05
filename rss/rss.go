package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Client struct {
	*http.Client
	Parser                *Parser
	CurrentRssCollections []*Rss
	CurrentRssItems       []*Item
	Sources               []string
}

// TODO add config
func NewClient() (*Client, error) {
	c := &http.Client{}
	return &Client{
		Client:                c,
		Parser:                &Parser{},
		Sources:               make([]string, 0, 100),
		CurrentRssCollections: make([]*Rss, 0, 100),
		CurrentRssItems:       make([]*Item, 0, 500),
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
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
		// println(string(bytes))
		rss, _ := c.xmlToRss(bytes)
		c.CurrentRssCollections = append(c.CurrentRssCollections, rss)
		c.CurrentRssItems = append(c.CurrentRssItems, rss.GetItems()...)
	}

	close(byteChan)

	for _, item := range c.CurrentRssItems {
		fmt.Println(item.Date + " " + item.Link)
		fmt.Println(len(c.CurrentRssItems))
	}
}

func (c *Client) xmlToRss(bytes []byte) (*Rss, error) {
	rss, _ := c.Parser.XmlToRss(bytes)
	return rss, nil
}

// Parse each xml block into an xml item of all the Rss we want to include
// Return a slice of rss items for each rss source (this should be dependent on the limit)
type Parser struct{}

func (p *Parser) XmlToRss(bytes []byte) (*Rss, error) {
	item := &Rss{}
	if err := xml.Unmarshal(bytes, &item); err != nil {
		panic(err)
	}
	return item, nil
}

// TODO: Replace this with an interface so that v1 items and v2 can return their Rss without the need for different structures
type Rss struct {
	XMLName xml.Name

	// v2
	Channel struct {
		Title string  `xml:"title"`
		Items []*Item `xml:"item"`
	} `xml:"channel,omitempty"`

	// v1 TODO: Figure out v1 structures
	Feed struct {
		Title string `xml:"title"`
	} `xml:",any,omitempty"`
}

// type RssInterperator {
// 	GetV2Items
// 	GetV1Items
// 	GetAtomItems
// }

// Add Filtering to this message
func (r *Rss) GetItems() []*Item {
	return r.Channel.Items
}

func (r *Rss) GetItemAtIndex(index int) *Item {
	if index < len(r.Channel.Items) {
		return &Item{}
	}
	return r.Channel.Items[index]
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Content     string `xml:"content"`
	Date        string `xml:"pubDate"`
}
