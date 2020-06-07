package search

import "testing"

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
			if want, got := token.typ, gotTokens[j].typ; want != got {
				t.Errorf("%d: token %d, want type %v, got type %v", i, j, want, got)
				continue
			}
			if want, got := token.content, gotTokens[j].content; want != got {
				t.Errorf("%d: token %d, want content %v, got content %v", i, j, want, got)
				continue
			}
		}
	}
}
