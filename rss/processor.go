package rss

import (
	"encoding/xml"
	"time"
)

type RssType uint8

const (
	ATOM RssType = iota
	V1
	V2
)

// Processor processess RSS XML from Atom and V2 Rss Feed Sources and returns an Rss Feed
// which is a slice Rss Items that was parsed from a given source
type Processor struct{}

// This returns an rss collection. This is a collection of parsed xml body data from a given source
func (p *Processor) XmlToRssCollection(bytes []byte) (*Collection, error) {
	converted := &rss{}

	if err := xml.Unmarshal(bytes, &converted); err != nil {
		panic(err)
	}

	collection := &Collection{
		Items: make([]*Item, 0, len(converted.Items)+len(converted.Channel.Items)),
	}

	for _, v2Item := range converted.Channel.Items {
		item := &Item{
			Type:        V2,
			Title:       v2Item.Title,
			Link:        v2Item.Link,
			Description: v2Item.Description,
			Content:     v2Item.Content,
			Date:        formatTime(v2Item.Date),
		}
		collection.Items = append(collection.Items, item)
	}

	for _, atomItem := range converted.Items {
		item := &Item{
			Type:        ATOM,
			Title:       atomItem.Title,
			Link:        atomItem.Link,
			Description: atomItem.Description,
			Date:        formatTime(atomItem.Updated),
		}

		if atomItem.Summary != "" {
			item.Content = atomItem.Summary
		}

		if atomItem.Content != "" {
			item.Content = atomItem.Content
		}

		if !time.Time(atomItem.Published).IsZero() {
			item.Date = formatTime(atomItem.Published)
		}

		collection.Items = append(collection.Items, item)
	}

	return collection, nil
}

// Normalized Rss Item
type Item struct {
	Type        RssType
	Date        formatTime
	Title       string
	Link        string
	Description string
	Content     string
}

type Collection struct {
	Items []*Item
}

// Add Filtering to this message
func (c *Collection) All() []*Item {
	return c.Items
}

func (c *Collection) ItemAtIndex(index int) *Item {
	items := c.All()

	if index < len(items) {
		return &Item{}
	}

	return items[index]
}

type formatTime time.Time

func (t *formatTime) String() string {
	return time.Time(*t).Format("01/02/2006")
}

// TODO: Replace this with an interface so that v1 items and v2 can return their Rss without the need for different structures
type rss struct {
	XMLName xml.Name

	// v2
	Channel struct {
		Title string `xml:"title"`
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			Content     string `xml:"content,omitempty"`
			Date        v2Time `xml:"pubDate,omitempty"`
		} `xml:"item"`
	} `xml:"channel,omitempty"`

	// Atom
	Items []struct {
		Title       string   `xml:"title"`
		Link        string   `xml:"id"`
		Description string   `xml:"description"`
		Content     string   `xml:"content"`
		Summary     string   `xml:"summary"`
		Published   atomTime `xml:"published"`
		Updated     atomTime `xml:"updated"`
	} `xml:"entry"`
}

type v2Time time.Time

func (t *v2Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// this format is the format the rss reader dates are being returned as
	const shortForm = time.RFC1123
	var v string

	d.DecodeElement(&v, &start)
	parse, err := time.Parse(shortForm, v)
	if err != nil {
		return err
	}

	*t = v2Time(parse)
	return nil
}

type atomTime time.Time

func (t *atomTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// this format is the format the rss reader dates are being returned as
	const shortForm = time.RFC3339
	var v string

	d.DecodeElement(&v, &start)
	parse, err := time.Parse(shortForm, v)
	if err != nil {
		return err
	}

	*t = atomTime(parse)
	return nil
}
