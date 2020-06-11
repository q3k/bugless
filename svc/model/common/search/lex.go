package search

import "fmt"

// lexer is a simple lexer/tokenizer/scanner for the bugless query language.
// It emits two types of tokens:
//  - word, ie a whitespace separated literal that's part of the query, that
//    can also be the result of grouping several query words into one query
//    word by wrapping them in "double quotes"
//  - colon, which is the ':' literal that's part of the query
//
// For example, the query:
//    title: foo "bar baz"bar : foo
// Would yield the following tokens:
//    word(title), colon, word(foo), word(bar baz), word(bar), colon, word(foo)
type lexer struct {
	s string
}

type token struct {
	// typ is the type of the token.
	typ tokenType
	// content is the literal, query-sourced content of the token.
	content string
}

type tokenType int

const (
	tokenTypeInvalid tokenType = iota
	// tokenWord is a word, either coming from a literal word in the query or
	// from a double quoted sentence of words including that can include
	// whitespaces.
	tokenWord
	// tokenColon is the literal ':' character that was part of the query (if
	// not quoted as part of a word).
	tokenColon
)

func (t token) String() string {
	switch t.typ {
	case tokenWord:
		return fmt.Sprintf("WORD<%q>", t.content)
	case tokenColon:
		return "COLON"
	}
	return "UNKNOWN"
}

func (l *lexer) read(n int) (string, bool) {
	if len(l.s) < n {
		return "", false
	}
	val := l.s[:n]
	l.s = l.s[n:]
	return val, true
}

func (l *lexer) lex() (tokens []token, terminated bool) {
	word := ""

	for {
		c, ok := l.read(1)
		if !ok {
			if word != "" {
				tokens = append(tokens, token{tokenWord, word})
				word = ""
			}
			terminated = true
			return
		}

		switch c {
		case ":":
			if word != "" {
				tokens = append(tokens, token{tokenWord, word})
				word = ""
			}
			tokens = append(tokens, token{tokenColon, c})
			continue
		case "\"":
			escaped := false
			if word != "" {
				tokens = append(tokens, token{tokenWord, word})
				word = ""
			}
			for {
				c, ok := l.read(1)
				if !ok {
					return
				}
				if c == "\\" && !escaped {
					escaped = true
					continue
				}
				if c == "\"" && !escaped {
					tokens = append(tokens, token{tokenWord, word})
					word = ""
					break
				}
				if escaped {
					escaped = false
				}
				word += c
			}
		case " ":
			fallthrough
		case "\t":
			if word != "" {
				tokens = append(tokens, token{tokenWord, word})
				word = ""
			}
		default:
			word += c
		}
	}
}
