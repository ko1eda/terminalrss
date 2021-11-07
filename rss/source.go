package rss

import "strings"

type SourceType uint8

const (
	HTTP SourceType = iota
	FILE
)

// Path can be the path to an xml file relative to the Rss Clients StorageRoot.
// It can also be an http source from a websites rss feed.
// Title is the name of the source or a nickname.
type Source struct {
	Type  SourceType
	Path  string
	Title string
}

// MapToSources converts a map of schmea path : title
// To a source of one of the above source types.
// Sources type are configured automatically,
// If the type is guessed incorrectly you can adjust its Type field manually.
func MapToSources(sources map[string]string) []*Source {
	s := make([]*Source, 0, len(sources))

	for source, title := range sources {
		temp, _ := NewSource(source, title, FILE)
		if strings.HasPrefix(strings.ToLower(source), "http") {
			temp.Path = source
			temp.Type = HTTP
		}
		s = append(s, temp)
	}

	return s
}

// Create a New Rss Source to add to the client.
// This can be a file path relative to the StorageRoot or an http source
func NewSource(path string, title string, sourceType SourceType) (*Source, error) {
	return &Source{Type: sourceType, Path: path, Title: title}, nil
}
