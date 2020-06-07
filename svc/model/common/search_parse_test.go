package common

import (
	"fmt"
	"testing"
)

func (n *nodeQuery) diff(o *nodeQuery) string {
	if want, got := len(n.elems), len(o.elems); want != got {
		return fmt.Sprintf("element count differ, want %+v, got %+v", n, o)
	}

	for i, w := range n.elems {
		g := o.elems[i]
		if diff := w.diff(&g); diff != "" {
			return fmt.Sprintf("element %d: %v", i, diff)
		}
	}
	return ""
}

func (n *nodeConstraintOrWord) diff(o *nodeConstraintOrWord) string {
	if n.constraint != nil {
		if o.constraint == nil {
			return fmt.Sprintf("want constraint, got word")
		}
		if diff := n.constraint.diff(o.constraint); diff != "" {
			return fmt.Sprintf("constraint: %v", diff)
		}
	}
	if n.word != nil {
		if o.word == nil {
			return fmt.Sprintf("want word, got constraint")
		}
		if diff := n.word.diff(o.word); diff != "" {
			return fmt.Sprintf("word: %v", diff)
		}
	}
	return ""
}

func (n *nodeConstraint) diff(o *nodeConstraint) string {
	if diff := n.key.diff(o.key); diff != "" {
		return fmt.Sprintf("key: %v", diff)
	}
	if diff := n.sep.diff(o.sep); diff != "" {
		return fmt.Sprintf("sep: %v", diff)
	}
	if diff := n.value.diff(o.value); diff != "" {
		return fmt.Sprintf("value: %v", diff)
	}
	return ""
}

func (n *nodeWord) diff(o *nodeWord) string {
	return n.word.diff(o.word)
}

func TestParse(t *testing.T) {
	for i, te := range []struct {
		tokens []token
		want   *nodeQuery
	}{
		{[]token{
			{tokenWord, "author"},
			{tokenColon, ":"},
			{tokenWord, "foo"},
		}, &nodeQuery{
			[]nodeConstraintOrWord{
				{constraint: &nodeConstraint{
					key:   token{tokenWord, "author"},
					sep:   token{tokenColon, ":"},
					value: token{tokenWord, "foo"},
				}},
			},
		}},
		{[]token{
			{tokenColon, ":"},
			{tokenWord, "author"},
			{tokenColon, ":"},
			{tokenWord, "foo"},
			{tokenWord, "bar baz"},
			{tokenWord, "title"},
			{tokenColon, ":"},
			{tokenWord, "foo"},
			{tokenWord, "bar"},
			{tokenWord, "baz"},
			{tokenColon, ":"},
		}, &nodeQuery{
			[]nodeConstraintOrWord{
				{constraint: &nodeConstraint{
					key:   token{tokenWord, "author"},
					sep:   token{tokenColon, ":"},
					value: token{tokenWord, "foo"},
				}},
				{word: &nodeWord{word: token{tokenWord, "bar baz"}}},
				{constraint: &nodeConstraint{
					key:   token{tokenWord, "title"},
					sep:   token{tokenColon, ":"},
					value: token{tokenWord, "foo"},
				}},
				{word: &nodeWord{word: token{tokenWord, "bar"}}},
				{word: &nodeWord{word: token{tokenWord, "baz"}}},
			},
		}},
	} {
		p := &parser{tokens: te.tokens}
		res := p.parse()
		if diff := te.want.diff(res); diff != "" {
			t.Errorf("test %d: %v", i, diff)
		}
	}
}
