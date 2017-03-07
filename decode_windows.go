// +build windows

package main

import (
	"encoding/xml"
	"os"
)

type BookmarksDocument struct {
	//XMLName xml.Name
	TopLevels []Bookmark `xml:"TreeViewNode"`
}

type Bookmark struct {
	//Typ    string `xml:"type,attr"`
	IsLeaf   bool
	Name     string
	RepoType string
	Path     string
	Children []Bookmark `xml:">TreeViewNode"`
}

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

func collectBookmarkLeaves(root Bookmark, dst *[]Bookmark) {
	if root.IsLeaf {
		*dst = append(*dst, root)
		return
	}
	for _, c := range root.Children {
		collectBookmarkLeaves(c, dst)
	}
}
