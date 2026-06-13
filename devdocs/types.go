package devdocs

// DocSet is a single documentation set available on DevDocs.io.
type DocSet struct {
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Release string `json:"release"`
	MTime   int64  `json:"mtime"`
	DbSize  int64  `json:"db_size"`
	Home    string `json:"home"`
	Code    string `json:"code"`
}

// Entry is a single documentation entry (function, class, method, etc.)
// within a doc set.
type Entry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

// EntryType is an aggregate of how many entries of a given type exist
// in a doc set.
type EntryType struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ─── wire types ───────────────────────────────────────────────────────────────

// wireLinks is the nested links object in the DevDocs API response.
type wireLinks struct {
	Home string `json:"home"`
	Code string `json:"code"`
}

// wireDocSet is the raw API shape for a documentation set.
type wireDocSet struct {
	Name    string    `json:"name"`
	Slug    string    `json:"slug"`
	Type    string    `json:"type"`
	Links   wireLinks `json:"links"`
	Version string    `json:"version"`
	Release string    `json:"release"`
	MTime   int64     `json:"mtime"`
	DbSize  int64     `json:"db_size"`
}

// wireIndex is the response body of /docs/{slug}/index.json.
type wireIndex struct {
	Entries []Entry     `json:"entries"`
	Types   []EntryType `json:"types"`
}

// wireToDocSet converts a wireDocSet to a DocSet, flattening the Links.
func wireToDocSet(w wireDocSet) DocSet {
	return DocSet{
		Name:    w.Name,
		Slug:    w.Slug,
		Type:    w.Type,
		Version: w.Version,
		Release: w.Release,
		MTime:   w.MTime,
		DbSize:  w.DbSize,
		Home:    w.Links.Home,
		Code:    w.Links.Code,
	}
}
