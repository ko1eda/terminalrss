package main

import (
	"fmt"
	"terminalrss/rss"
)

func main() {
	client, _ := rss.NewClient()
	client.AddSources(
		rss.MapToSources(
			map[string]string{
				"https://xkcd.com/atom.xml":                                                "xkcd",
				"https://www.theverge.com/google/rss/index.xml":                            "The Verge",
				"https://www.gobeyond.dev/rss/":                                            "Go Beyond Dev Blog",
				"https://www.reutersagency.com/feed/?taxonomy=best-sectors&post_type=best": "Reuters",
				"https://www.latimes.com/local/rss2.0.xml":                                 "LA TIMES",
			},
		))

	feed := client.Load(client.ListSources())
	// fmt.Printf("%p %p", feed, client.Feed())
	for _, item := range feed {
		fmt.Printf("\n\n%s \n%s \n%s \n%s \n%s \n\n", item.Source.Title, item.Date.String(), item.Title, item.Creator, item.Description)
	}
	fmt.Println(feed.GetSize())
}
