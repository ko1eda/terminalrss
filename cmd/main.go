package main

import "terminalrss"

func main() {
	client, _ := terminalrss.NewClient()
	client.AddSources(
		[]string{
			"https://xkcd.com/atom.xml",
			"https://www.gobeyond.dev/rss/",
			"https://www.reutersagency.com/feed/?taxonomy=best-sectors&post_type=best",
		},
	)
	client.Refresh()
	//TODO build tui
}
