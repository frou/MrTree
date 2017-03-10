package main

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

func collectBookmarkLeaves(root Bookmark, dst *[]Bookmark) {
	if root.IsLeaf {
		*dst = append(*dst, root)
		return
	}
	for _, c := range root.Children {
		collectBookmarkLeaves(c, dst)
	}
}
