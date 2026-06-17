package opml

import (
	"encoding/xml"
	"fmt"
	"io"
	"sort"
)

func Export(specs []FeedSpec, w io.Writer) error {
	byFolder := make(map[string][]FeedSpec)
	var noFolder []FeedSpec
	for _, s := range specs {
		if s.Folder == "" {
			noFolder = append(noFolder, s)
		} else {
			byFolder[s.Folder] = append(byFolder[s.Folder], s)
		}
	}

	folderNames := make([]string, 0, len(byFolder))
	for name := range byFolder {
		folderNames = append(folderNames, name)
	}
	sort.Strings(folderNames)

	var bodyOutlines []opmlOutline

	for _, name := range folderNames {
		feeds := byFolder[name]
		var children []opmlOutline
		for _, f := range feeds {
			children = append(children, opmlOutline{
				Text:   f.Title,
				Type:   "rss",
				XMLURL: f.URL,
			})
		}
		bodyOutlines = append(bodyOutlines, opmlOutline{
			Text:     name,
			Outlines: children,
		})
	}

	for _, f := range noFolder {
		bodyOutlines = append(bodyOutlines, opmlOutline{
			Text:   f.Title,
			Type:   "rss",
			XMLURL: f.URL,
		})
	}

	doc := opmlDoc{
		Version: "2.0",
		Head:    opmlHead{Title: "Feed Tracker Export"},
		Body:    opmlBody{Outlines: bodyOutlines},
	}

	output, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal opml: %w", err)
	}

	header := []byte(xml.Header)
	if _, err := w.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	if _, err := w.Write(output); err != nil {
		return fmt.Errorf("write body: %w", err)
	}
	_, err = w.Write([]byte("\n"))
	return err
}
