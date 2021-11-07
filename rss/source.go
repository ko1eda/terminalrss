package rss

import "strings"

type SourceType uint8

const (
	HTTP SourceType = iota
	FILE
)

// Path can be the path to an xml file relative to the project root
// It can also be an http source from a websites rss feed
type Source struct {
	Type SourceType
	Path string
}

// Create a New Rss Source to add to the client.
// This can be a file path relative to the project root or an http source
func NewSource(p string, t SourceType) (*Source, error) {
	return &Source{Type: t, Path: p}, nil
}

// StringToSource converts a list of string sources to Rss sources
// It will automatically determine if the link is an http source or a file
func StringsToSources(sources []string) []*Source {
	s := make([]*Source, 0, len(sources))

	for _, source := range sources {
		temp, _ := NewSource(source, FILE)
		if strings.HasPrefix(source, "http") {
			temp.Path = source
			temp.Type = HTTP
		}
		s = append(s, temp)
	}

	return s
}
