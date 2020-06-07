package search

import "fmt"

type token struct {
	typ     tokenType
	content string
}

type tokenType int

const (
	tokenTypeInvalid tokenType = iota
	tokenWord
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

type lexer struct {
	s       string
	peeking int
}

func (l *lexer) peek(n int) (string, bool) {
	if len(l.s) < (l.peeking + n) {
		return "", false
	}
	val := l.s[l.peeking : l.peeking+n]
	l.peeking += n
	return val, true
}

func (l *lexer) commit() {
	l.s = l.s[l.peeking:]
	l.peeking = 0
}

func (l *lexer) read(n int) (string, bool) {
	if l.peeking != 0 {
		panic("read with uncommited peek")
	}
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
		c, ok := l.peek(1)
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
			l.commit()
			if word != "" {
				tokens = append(tokens, token{tokenWord, word})
				word = ""
			}
			tokens = append(tokens, token{tokenColon, c})
			continue
		case "\"":
			l.commit()
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
			l.commit()
			if word != "" {
				tokens = append(tokens, token{tokenWord, word})
				word = ""
			}
		default:
			l.commit()
			word += c
		}
	}
}
