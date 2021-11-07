package rss

import "strings"

type SourceType uint8

const (
	HTTP SourceType = iota
	FILE
)

// SourceMapper is a source management component.
// It maps file sources and http sources to a title and source type
// It allows us to perform List and Find operations on those sources
type SourceMapper struct {
	HasSources  bool
	SourceSlice []*Source
	Sources     map[string]*Source
}

// Path can be the path to an xml file relative to the Rss Clients StorageRoot.
// It can also be an http source from a websites rss feed.
// Title is the name of the source or a nickname.
type Source struct {
	Type  SourceType
	Path  string
	Title string
}

// Create and new SourceMapper with sensible defaults
func NewSourceMapper() (*SourceMapper, error) {
	return &SourceMapper{
		SourceSlice: make([]*Source, 0, 100),
		Sources:     make(map[string]*Source, 100),
		HasSources:  false,
	}, nil
}

// Create a New Rss Source to add to the client.
// This can be a file path relative to the StorageRoot or an http source
func NewSource(path string, title string, sourceType SourceType) (*Source, error) {
	return &Source{Type: sourceType, Path: path, Title: title}, nil
}

// MapToSources is a convenience function that converts a map of schmea path : title
// To a source of one of the above source types.
// Sources type are configured automatically.
// You can use this in conjunction with AddSources or RemoveSources
// to easily create multiple sources from a map of strings.
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

// Add sources to client to load Rss data from
func (s *SourceMapper) AddSources(sources []*Source) {
	m := s.Sources

	if !s.HasSources {
		m = make(map[string]*Source, 100)
		s.HasSources = true
	}

	for _, source := range sources {
		_, hit := m[source.Path]

		if hit {
			continue
		}
		// If the source isn't in our map of sources
		// creat a new rss item slice that we will
		// use when we load rss items
		s.SourceSlice = append(s.SourceSlice, source)
		m[source.Path] = source
	}

	s.Sources = m
}

// Removes any number of sources from the client
func (s *SourceMapper) RemoveSources(sources []*Source) {
	for _, source := range sources {
		_, hit := s.Sources[source.Path]

		if !hit {
			continue
		}

		delete(s.Sources, source.Path)
	}

	// If someone passed in a nil slice or empty slice
	// there is no need to recreate the source slice
	if len(sources) < 1 {
		return
	}
	// Here we replace source slices with whatever sources are left in the map
	s.SourceSlice = make([]*Source, 0, len(s.Sources))

	for _, source := range s.Sources {
		s.SourceSlice = append(s.SourceSlice, source)
	}
}

// Returns a slice of all Rss sources the Client contains
func (s *SourceMapper) ListSources() []*Source {
	return s.SourceSlice
}

// Find a given source by path or title
// Returns true if we have hit and miss with an empty source if we don't have a match
func (s *SourceMapper) FindSource(target string) (*Source, bool) {
	hit := false
	source := &Source{}
	for path := range s.Sources {
		if target == path || target == s.Sources[path].Title {
			hit = true
			source = s.Sources[path]
		}
	}

	if !hit {
		return source, hit
	}

	return source, hit
}
