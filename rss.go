package terminalrss

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
	Parser       *Parser
	CurrentItems []*Item
	Sources      []string
}

// TODO add config
func NewClient() (*Client, error) {
	c := &http.Client{}
	return &Client{
		Client:       c,
		Parser:       &Parser{},
		Sources:      make([]string, 0, 100),
		CurrentItems: make([]*Item, 0, 500),
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel() // TODO: look up defer

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
			continue
		}
		c.parse(bytes)
	}

	close(byteChan)

	for _, item := range c.CurrentItems {
		fmt.Println(item.Date + " " + item.Description)
	}
}

func (c *Client) parse(bytes []byte) error {
	rssContent, _ := c.Parser.ParseXMLBytes(bytes)
	c.CurrentItems = append(c.CurrentItems, rssContent.GetItemCollection()...)
	return nil
}

// Parse each xml block into an xml item of all the data we want to include
// Return a slice of rss items for each rss source (this should be dependent on the limit)
type Parser struct{}

func (p *Parser) ParseXMLBytes(bytes []byte) (*Content, error) {
	item := &Content{}
	if err := xml.Unmarshal(bytes, &item); err != nil {
		panic(err)
	}
	return item, nil
}

// TODO: Replace this with an interface so that v1 items and v2 can return their data without the need for different structures
type Content struct {
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

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Content     string `xml:"content"`
	Date        string `xml:"pubDate"`
}

func (c *Content) GetItemCollection() []*Item {
	return c.Channel.Items
}
