package main

import "terminalrss/rss"

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
	client.Refresh()
	//TODO build tui
}
