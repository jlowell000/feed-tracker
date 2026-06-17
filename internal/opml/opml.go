package opml

import (
	"encoding/xml"
	"fmt"
	"os"
)

type FeedSpec struct {
	URL    string
	Title  string
	Folder string
}

type opmlDoc struct {
	XMLName xml.Name  `xml:"opml"`
	Version string    `xml:"version,attr"`
	Head    opmlHead  `xml:"head"`
	Body    opmlBody  `xml:"body"`
}

type opmlHead struct {
	Title string `xml:"title"`
}

type opmlBody struct {
	Outlines []opmlOutline `xml:"outline"`
}

type opmlOutline struct {
	Text     string        `xml:"text,attr,omitempty"`
	Title    string        `xml:"title,attr,omitempty"`
	Type     string        `xml:"type,attr,omitempty"`
	XMLURL   string        `xml:"xmlUrl,attr,omitempty"`
	HTMLURL  string        `xml:"htmlUrl,attr,omitempty"`
	Outlines []opmlOutline `xml:"outline"`
}

func ParseFile(path string) ([]FeedSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var doc opmlDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse xml: %w", err)
	}

	var specs []FeedSpec
	for _, o := range doc.Body.Outlines {
		walkOutline(o, "", &specs)
	}

	return specs, nil
}

func walkOutline(o opmlOutline, folder string, specs *[]FeedSpec) {
	if o.XMLURL != "" {
		title := o.Title
		if title == "" {
			title = o.Text
		}
		*specs = append(*specs, FeedSpec{
			URL:    o.XMLURL,
			Title:  title,
			Folder: folder,
		})
		return
	}

	name := o.Text
	if name == "" {
		name = o.Title
	}
	if name == "" {
		name = folder
	}

	for _, child := range o.Outlines {
		walkOutline(child, name, specs)
	}
}
