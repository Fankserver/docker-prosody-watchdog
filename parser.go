package main

import (
	"encoding/xml"

	"github.com/mmcdole/gofeed"
)

type table struct {
	Tr []tableRow `xml:"tr"`
}
type tableRow struct {
	Th string `xml:"th"`
	Td string `xml:"td"`
}
type content struct {
	ChangeSet   string
	Branch      string
	Bookmark    string
	Tag         string
	User        string
	Description string
	Files       string
}

func parseContent(item *gofeed.Item) (*content, error) {
	var t table
	err := xml.Unmarshal([]byte(item.Description), &t)
	if err != nil {
		return nil, err
	}

	return &content{
		ChangeSet:   t.Tr[0].Td,
		Branch:      t.Tr[1].Td,
		Bookmark:    t.Tr[2].Td,
		Tag:         t.Tr[3].Td,
		User:        t.Tr[4].Td,
		Description: t.Tr[5].Td,
		Files:       t.Tr[6].Td,
	}, nil
}
