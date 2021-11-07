package main

import (
	"fmt"
	"terminalrss/rss"
)

func main() {
	client, _ := rss.NewClient()
	client.AddSources(rss.StringsToSources(
		[]string{
			"https://xkcd.com/atom.xml",
			"https://www.theverge.com/google/rss/index.xml",
			"https://www.gobeyond.dev/rss/",
			"https://www.reutersagency.com/feed/?taxonomy=best-sectors&post_type=best",
		},
	))
	feed := client.Load(client.ListSources())

	// fmt.Printf("%p %p", feed, client.Feed())
	for _, item := range feed {
		fmt.Printf("\n\n%s \n%s \n%s \n %s \n\n", item.Date.String(), item.Title, item.Creator, item.Description)
	}
}
