package common

import (
	"fmt"
	"testing"
)

func (t token) diff(o token) string {
	if want, got := t.typ, o.typ; want != got {
		return fmt.Sprintf("want type %v, got type %v", want, got)
	}
	if want, got := t.content, o.content; want != got {
		return fmt.Sprintf("want content %q, got content %q", want, got)
	}
	return ""
}

func TestLex(t *testing.T) {
	for i, te := range []struct {
		s          string
		terminated bool
		want       []token
	}{
		{"foo bar", true, []token{{tokenWord, "foo"}, {tokenWord, "bar"}}},
		{" 	foo 	 bar	", true, []token{{tokenWord, "foo"}, {tokenWord, "bar"}}},
		{"  \"bug less\" bug less ", true, []token{
			{tokenWord, "bug less"}, {tokenWord, "bug"}, {tokenWord, "less"},
		}},
		{"error \"\\\"foo\\\" not defined\"", true, []token{
			{tokenWord, "error"}, {tokenWord, "\"foo\" not defined"},
		}},
		{"author:q3k foo bar baz", true, []token{
			{tokenWord, "author"}, {tokenColon, ":"}, {tokenWord, "q3k"},
			{tokenWord, "foo"}, {tokenWord, "bar"}, {tokenWord, "baz"},
		}},
		{"title:\"bug less\" author:q3k", true, []token{
			{tokenWord, "title"}, {tokenColon, ":"}, {tokenWord, "bug less"},
			{tokenWord, "author"}, {tokenColon, ":"}, {tokenWord, "q3k"},
		}},
	} {
		l := &lexer{s: te.s}
		gotTokens, gotTerminated := l.lex()
		if want, got := te.terminated, gotTerminated; want != got {
			t.Errorf("%d: wanted terminated %v, got %v", i, want, got)
			continue
		}

		if want, got := len(te.want), len(gotTokens); want != got {
			t.Errorf("%d: token count mismatch, want %v, got %v", i, te.want, gotTokens)
			continue
		}

		for j, token := range te.want {
			if diff := token.diff(gotTokens[j]); diff != "" {
				t.Errorf("%d: token %d, %v", i, j, diff)
				continue
			}
		}
	}
}
