package rss

import (
	"sort"
	"time"
)

type SortOrder uint8

const (
	DATE_ASC SortOrder = iota
	DATE_DSC
)

// A slice of Converted RSS Items
type Feed []*Item

func (f Feed) Len() int      { return len(f) }
func (f Feed) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

// Less returns true if the first value is less than the second. so we use Date Before here
func (f Feed) Less(i, j int) bool { return time.Time(f[i].Date).Before(time.Time(f[j].Date)) }

// Sort Feed by date using SortOrder constants
func (f *Feed) SortBy(sorter SortOrder) {
	switch sorter {
	case DATE_ASC:
		sort.Stable(f)
	case DATE_DSC:
		sort.Stable(sort.Reverse(f))
	}
}
