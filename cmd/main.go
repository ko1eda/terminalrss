package main

import (
	"fmt"
	"terminalrss/rss"
)

func main() {
	client, _ := rss.NewClient()
	client.AddSources(
		[]string{
			"https://xkcd.com/atom.xml",
			"https://www.theverge.com/google/rss/index.xml",
			"https://www.gobeyond.dev/rss/",
			"https://www.reutersagency.com/feed/?taxonomy=best-sectors&post_type=best",
		},
	)
	// Remove sources 1 and 2
	client.RemoveSources(client.ListSources()[1:3])

	// Load sources 0 and 3
	feed := client.Load(client.ListSources())

	for _, item := range feed {
		fmt.Printf("\n\n%s \n%s \n%s \n\n", item.Date.String(), item.Title, item.Description)
	}
}
