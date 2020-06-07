package search

import "strings"

// Bugless provides a query language for search queries.
//
// Currently it's very simplistic, and made up of:
//  - Key/value filters, like "author:q3k"
//  - Keywords
// A query like 'author:q3k foo bar "bar foo" status:open' would for example
// get parsed as 'all issues authored by q3k AND whose status is open AND to be
// ordered by relevancy according to the keywords 'foo', 'bar' and 'bar foo'.
//
// Keywords are defined as words that are part of the issue title, body or other fields.
//
// TODO(q3k): document this once this stabilizes/evolves.
//
// TODO(q3k): define a formal grammar before doing any more work on this.

// Search query is a parsed, but lightly typed search query from the user. It
// is the result of a string-based query (like creator:q3k foo status:bar).
// All search string elements that could have been extracted end up in the
// 'key' fields.
type SearchQuery struct {
	// Key fields. If set, the query selects that the given key must be set to
	// its corresponding value.
	ID       string
	Creator  string
	Assignee string
	Status   string

	// All words that are not part of key/value filters.
	Keywords []string
	// The original query.
	OriginalQuery string
}

// ParseSearch parses the given string as a search query and returns a
// SearchQuery object that is a semi-raw representation of the query: ie., with
// fields names detected, but not type checked. The consumer of this type can
// consume those values in a strict or fuzzy way depending on whether they are
// valid or not.
func ParseSearch(s string) *SearchQuery {
	res := &SearchQuery{
		OriginalQuery: s,
	}

	l := lexer{s: s}
	tokens, _ := l.lex()
	p := parser{tokens: tokens}
	q := p.parse()

	for _, el := range q.elems {
		if el.constraint != nil {
			switch strings.ToLower(el.constraint.key.content) {
			case "id":
				res.ID = el.constraint.value.content
			case "creator":
				res.Creator = el.constraint.value.content
			case "assignee":
				res.Assignee = el.constraint.value.content
			case "status":
				res.Status = el.constraint.value.content
			}
		}
		if el.word != nil {
			res.Keywords = append(res.Keywords, el.word.word.content)
		}
	}
	return res
}
