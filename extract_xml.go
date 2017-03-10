// +build windows

package main

import (
	"encoding/xml"
	"os"
)

// Decode bookmarks file in the Windows format (XML)
func decodeBookmarksFile(path string) ([]Bookmark, error) {
	bookmarksFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer bookmarksFile.Close()

	decoder := xml.NewDecoder(bookmarksFile)
	doc := new(BookmarksDocument)
	if err := decoder.Decode(doc); err != nil {
		return nil, err
	}

	var marks []Bookmark
	synthRoot := Bookmark{Children: doc.TopLevels}
	collectBookmarkLeaves(synthRoot, &marks)
	return marks, nil
}
