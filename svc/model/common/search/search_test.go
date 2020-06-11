package search

import (
	"fmt"
	"testing"
)

func (q *Query) diff(o *Query) string {
	if want, got := q.ID, o.ID; want != got {
		return fmt.Sprintf("wanted ID %q, got %q", want, got)
	}
	if want, got := q.Author, o.Author; want != got {
		return fmt.Sprintf("wanted Author %q, got %q", want, got)
	}
	if want, got := q.Assignee, o.Assignee; want != got {
		return fmt.Sprintf("wanted Assignee %q, got %q", want, got)
	}
	if want, got := q.Status, o.Status; want != got {
		return fmt.Sprintf("wanted Status %q, got %q", want, got)
	}
	if want, got := len(q.Keywords), len(o.Keywords); want != got {
		return fmt.Sprintf("wanted Keywords %v got %v", want, got)
	}
	for i, w := range o.Keywords {
		g := o.Keywords[i]
		if w != g {
			return fmt.Sprintf("keword %d: wanted %q, got %q", i, w, g)
		}
	}
	return ""
}

func TestParseSearch(t *testing.T) {
	for i, te := range []struct {
		s    string
		want *Query
	}{
		{"assignee:q3k status:assigned", &Query{
			Assignee: "q3k", Status: "assigned",
		}},
		{"id:1234", &Query{
			ID: "1234",
		}},
		{"bugless \"bug less\"", &Query{
			Keywords: []string{"bugless", "bug less"},
		}},
		{"author:\"q3k@q3k.org\" \"foo bar\"", &Query{
			Author:   "q3k@q3k.org",
			Keywords: []string{"foo bar"},
		}},
	} {
		got := ParseSearch(te.s)
		if diff := te.want.diff(got); diff != "" {
			fmt.Errorf("test %d: %v", i, diff)
		}
	}
}
