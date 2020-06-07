package common

// parser for bugless query language.
type parser struct {
	tokens []token
}

func (p *parser) peek(n int) ([]token, bool) {
	if len(p.tokens) < n {
		return nil, false
	}
	return p.tokens[:n], true
}

func (p *parser) read(n int) ([]token, bool) {
	val, ok := p.peek(n)
	if ok {
		p.tokens = p.tokens[n:]
	}
	return val, ok
}

// nodeQuery is the top-level AST node of the query.
// It is made up of elements of the query, that can be a constraint (a
// key:value pair) or a word (a word that's not part of a key:value pair).
type nodeQuery struct {
	elems []nodeConstraintOrWord
}

// nodeConstraint is a constraint or word. One and only one field must be set.
type nodeConstraintOrWord struct {
	constraint *nodeConstraint
	word       *nodeWord
}

// nodeConstraint is a key:value filter in the query.
type nodeConstraint struct {
	// key is the part before the colon, ie. the field name.
	key token
	// sep is the separator/predicate of the constraint, currently only ':'
	// (equality).
	sep token
	// value is the part after che colon, ie. the field value filter.
	value token
}

// node word is a free-standing word (or "double quoted" group of literal
// words) that's not part of any key:value constraint.
type nodeWord struct {
	word token
}

func (p *parser) parse() *nodeQuery {
	res := &nodeQuery{}
	for {
		constraint := p.parseConstraint()
		if constraint != nil {
			res.elems = append(res.elems, nodeConstraintOrWord{constraint: constraint})
			continue
		}

		word := p.parseWord()
		if word != nil {
			res.elems = append(res.elems, nodeConstraintOrWord{word: word})
			continue
		}

		// last token not even a word, just ignore it.
		_, ok := p.read(1)
		if !ok {
			break
		}
	}
	return res
}

func (p *parser) parseConstraint() *nodeConstraint {
	toks, ok := p.peek(3)
	if !ok {
		return nil
	}
	if toks[0].typ != tokenWord || toks[1].typ != tokenColon || toks[2].typ != tokenWord {
		return nil
	}
	p.read(3)
	return &nodeConstraint{
		key:   toks[0],
		sep:   toks[1],
		value: toks[2],
	}
}

func (p *parser) parseWord() *nodeWord {
	toks, ok := p.peek(1)
	if !ok {
		return nil
	}
	if toks[0].typ != tokenWord {
		return nil
	}
	p.read(1)
	return &nodeWord{
		word: toks[0],
	}
}
