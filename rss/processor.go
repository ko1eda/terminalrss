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

// This returns an rss feed. This is a feed of parsed xml body data from a given source
func (p *Processor) XmlToRssFeed(bytes []byte) (Feed, error) {
	converted := &rss{}

	if err := xml.Unmarshal(bytes, &converted); err != nil {
		panic(err)
	}

	feed := Feed(make([]*Item, 0, len(converted.Items)+len(converted.Channel.Items)))

	for _, v2Item := range converted.Channel.Items {
		item := &Item{
			Type:        V2,
			Title:       v2Item.Title,
			Link:        v2Item.Link,
			Description: v2Item.Description,
			Content:     v2Item.Content,
			Creator:     v2Item.Creator,
			Date:        formatTime(v2Item.Date),
		}
		feed = append(feed, item)
	}

	for _, atomItem := range converted.Items {
		item := &Item{
			Type:        ATOM,
			Title:       atomItem.Title,
			Link:        atomItem.Link,
			Description: atomItem.Description,
			Creator:     atomItem.Creator,
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

		feed = append(feed, item)
	}

	return feed, nil
}

// Normalized Rss Item
type Item struct {
	Source      *Source
	Type        RssType
	Date        formatTime
	Title       string
	Link        string
	Description string
	Content     string
	Creator     string
}

type formatTime time.Time

func (t *formatTime) String() string {
	return time.Time(*t).Format("01/02/2006")
}

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
			Creator     string `xml:"creator"`
			Date        v2Time `xml:"pubDate,omitempty"`
		} `xml:"item"`
	} `xml:"channel,omitempty"`

	// Atom
	Items []struct {
		Title       string   `xml:"title"`
		Link        string   `xml:"id"`
		Description string   `xml:"description"`
		Content     string   `xml:"content"`
		Creator     string   `xml:"author>name"`
		Summary     string   `xml:"summary"`
		Published   atomTime `xml:"published"`
		Updated     atomTime `xml:"updated"`
	} `xml:"entry"`
}

type v2Time time.Time

func (t *v2Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	const shortForm = time.RFC1123
	var v string

	d.DecodeElement(&v, &start)
	parse, err := time.Parse(shortForm, v)

	// Try another time format without the prefix on the day of the month
	if err != nil {
		// TODO: LOG ERROR
		parse, _ = time.Parse("Mon, _2 Jan 2006 15:04:05 MST", v)
	}

	*t = v2Time(parse)

	return nil
}

type atomTime time.Time

func (t *atomTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
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
