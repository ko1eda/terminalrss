package rss_test

import (
	"terminalrss/rss"
	"testing"
)

func Test_Sources(t *testing.T) {
	client, _ := rss.NewClient()
	client.AddSources(
		rss.MapToSources(
			map[string]string{
				"https://xkcd.com/atom.xml":                     "xkcd",
				"https://www.theverge.com/google/rss/index.xml": "The Verge",
			},
		))

	t.Run("List_Sources", func(t *testing.T) {
		var want, got int
		want = 2
		got = len(client.ListSources())
		if want != got {
			t.Errorf("want: %v , got: %v", want, got)
		}
	})

	t.Run("Find_Source", func(t *testing.T) {
		var want, got bool

		_, got = client.FindSource("The Blerge")
		if want != got {
			t.Errorf("want: %v , got: %v", want, got)
		}

		want = true
		_, got = client.FindSource("The Verge")
		if want != got {
			t.Errorf("want: %v , got: %v", want, got)
		}
	})

	t.Run("Remove_Sources", func(t *testing.T) {
		client, _ := rss.NewClient()
		client.AddSources(
			rss.MapToSources(
				map[string]string{
					"https://xkcd.com/atom.xml":                     "xkcd",
					"https://www.theverge.com/google/rss/index.xml": "The Verge",
				},
			))

		var want, got int
		want = 0
		client.RemoveSources(client.ListSources()[:])
		got = len(client.ListSources())
		if want != got {
			t.Errorf("want: %v , got: %v", want, got)
		}
	})

}
