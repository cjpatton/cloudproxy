// Copyright (c) 2014, Kevin Walsh.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This code borrows heavily from the lexer design and implementation for the
// template package. See http://golang.org/src/pkg/text/template/parse/parse.go

package auth

import (
	"encoding/base64"
	"fmt"
)

// The functions in this file use one token lookahead, but only when more input
// is actually called for. The lexer may read one rune ahead while getting a
// token, but will unread that rune when the token is completed. The goal is to
// allow parsing an element out of a string or input stream that contains other
// data after the element.
//
// The parseX() functions properly handle outer parenthesis. For
// example, parsePred() will accept "P(1)", "(P(1))", and " ( ((P((1 )) ) ))".
// The expectX() functions do not allow outer parenthesis. So
// expectPred() will handle "P(1)" and "P( (( 1) ))", but not "(P(1))".
//
// Onless otherwise documented, in all cases the parseX() and expectX()
// functions are greedy, consuming input until either an error is encountered or
// the element can't be expanded further.

// parser holds the state of the recursive descent parser.
type parser struct {
	lex           *lexer
	lookahead     token
	haveLookahead bool
}

func (parser *parser) cur() token {
	if !parser.haveLookahead {
		parser.lookahead = parser.lex.nextToken()
		parser.haveLookahead = true
	}
	return parser.lookahead
}

// advance discards lookahead; the next call to cur() will get a new token.
func (parser *parser) advance() {
	parser.haveLookahead = false
}

// expect checks whether cur matches t and, if so, advances to the next token.
func (parser *parser) expect(t token) error {
	if parser.cur() != t {
		return fmt.Errorf("expected %q, found %v", t.val, parser.cur())
	}
	parser.advance()
	return nil
}

// skipOpenParens skips and counts open parens.
func (parser *parser) skipOpenParens() int {
	n := 0
	for parser.cur() == tokenLP {
		parser.advance()
		n++
	}
	return n
}

// expectCloseParens expects n close parens.
func (parser *parser) expectCloseParens(n int) error {
	for n > 0 {
		err := parser.expect(tokenRP)
		if err != nil {
			return err
		}
		n--
	}
	return nil
}

// expectPrin expects a Prin.
func (parser *parser) expectPrin() (p Prin, err error) {
	if parser.cur() != tokenTPM && parser.cur() != tokenKey {
		err = fmt.Errorf(`expected "key" or "tpm", found %v`, parser.cur())
		return
	}
	p.Type = parser.cur().val.(string)
	parser.advance()
	if r := parser.lex.peek(); r != '(' {
		err = fmt.Errorf(`expected '(' directly after "key", found %q`, r)
		return
	}
	err = parser.expect(tokenLP)
	if err != nil {
		return
	}
	key, err := parser.parseStr()
	if err != nil {
		return
	}
	err = parser.expect(tokenRP)
	if err != nil {
		return
	}
	p.Key, err = base64.URLEncoding.DecodeString(string(key))
	if err != nil {
		return
	}
	for parser.lex.peek() == '.' {
		p.Ext, err = expectSubPrin()
	}
	return
}

// parsePrin parses a Prin with optional outer parens.
func (parser *parser) parsePrin() (p Prin, err error) {
	n := parser.skipOpenParens()
	p, err = parser.expectPrin()
	if err != nil {
		return
	}
	err = parser.expectCloseParens(n)
	return
}

// expectSubPrin expects a SubPrin.
func (parser *parser) expectSubPrin() (s SubPrin, err error) {
	if parser.cur() != tokenDot {
		err = fmt.Errorf(`expected '.', found %v`, parser.cur())
		return
	}
	parser.advance()
	name, args, err := parser.expectNameAndArgs()
	if err != nil {
		return
	}
	s = append(s, PrinExt{name, args})
	for parser.lex.peek() == '.' {
		parser.advance()
		name, args, err = parser.expectNameAndArgs()
		if err != nil {
			return
		}
		s = append(s, PrinExt{name, args})
	}
	return
}

// expectNameAndArgs expects an identifier, optionally followed by
// a parenthesized list of zero or more comma-separated terms.
func (parser *parser) expectNameAndArgs() (string, []Term, error) {
	if parser.cur().typ != itemIdentifier {
		return "", nil, fmt.Errorf("expected identifier, found %v", parser.cur())
	}
	name := parser.cur().val.(string)
	parser.advance()
	if parser.lex.peek() != '(' {
		// no parens
		return name, nil, nil
	}
	if parser.cur() != tokenLP {
		panic("not reached")
	}
	parser.advance()
	if parser.cur() == tokenRP {
		// empty parens
		parser.advance()
		return name, nil, nil
	}
	var args []Term
	for {
		t, err := parser.parseTerm()
		if err != nil {
			return "", nil, err
		}
		args = append(args, t)
		if parser.cur() != tokenComma {
			break
		}
		parser.advance()
	}
	err := parser.expect(tokenRP)
	if err != nil {
		return "", nil, err
	}
	return name, args, nil
}

// expectStr expects a Str.
func (parser *parser) expectStr() (Str, error) {
	if parser.cur().typ != itemStr {
		return "", fmt.Errorf("expected string, found %v", parser.cur())
	}
	t := Str(parser.cur().val.(string))
	parser.advance()
	return t, nil
}

// parseStr parses a Str with optional outer parens.
func (parser *parser) parseStr() (t Str, err error) {
	n := parser.skipOpenParens()
	t, err = parser.expectStr()
	if err != nil {
		return
	}
	err = parser.expectCloseParens(n)
	return
}

// expectInt expects an Int.
func (parser *parser) expectInt() (Int, error) {
	if parser.cur().typ != itemInt {
		return 0, fmt.Errorf("expected int, found %v", parser.cur())
	}
	t := Int(parser.cur().val.(int64))
	parser.advance()
	return t, nil
}

// parseInt parses an Int with optional outer parens.
func (parser *parser) parseInt() (Int, error) {
	n := parser.skipOpenParens()
	t, err := parser.expectInt()
	if err != nil {
		return 0, err
	}
	err = parser.expectCloseParens(n)
	if err != nil {
		return 0, err
	}
	return t, nil
}

// expectTerm expects a Term.
func (parser *parser) expectTerm() (Term, error) {
	switch parser.cur().typ {
	case itemStr:
		return parser.expectStr()
	case itemInt:
		return parser.expectInt()
	case itemKeyword:
		return parser.expectPrin()
	default:
		return nil, fmt.Errorf("expected term, found %v", parser.cur())
	}
}

// parseTerm parses a Term with optional outer parens.
func (parser *parser) parseTerm() (Term, error) {
	n := parser.skipOpenParens()
	t, err := parser.expectTerm()
	if err != nil {
		return nil, err
	}
	err = parser.expectCloseParens(n)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// expectPred expects a Pred.
func (parser *parser) expectPred() (f Pred, err error) {
	name, args, err := parser.expectNameAndArgs()
	if err != nil {
		return
	}
	return Pred{name, args}, nil
}

// parsePred parses a Pred with optional outer parens.
func (parser *parser) parsePred() (f Pred, err error) {
	n := parser.skipOpenParens()
	f, err = parser.expectPred()
	if err != nil {
		return
	}
	err = parser.expectCloseParens(n)
	return
}

// expectConst expects a Const.
func (parser *parser) expectConst() (f Const, err error) {
	if parser.cur() != tokenTrue && parser.cur() != tokenFalse {
		err = fmt.Errorf("expected Const, found %v", parser.cur())
		return
	}
	f = Const(parser.cur() == tokenTrue)
	parser.advance()
	return
}

// parseConst parses a Const with optional outer parens.
func (parser *parser) parseConst() (f Const, err error) {
	n := parser.skipOpenParens()
	f, err = parser.expectConst()
	if err != nil {
		return
	}
	err = parser.expectCloseParens(n)
	return
}

// expectFrom optionally expects a "(from|until) int" clause for a says formula.
func (parser *parser) expectOptionalTime(t token) (*int64, error) {
	if parser.cur() != t {
		return nil, nil
	}
	parser.advance()
	i, err := parser.parseInt()
	if err != nil {
		return nil, err
	}
	val := int64(i)
	return &val, nil
}

// expectSaysOrSpeaksfor expects a says or speaksfor formula. If greedy is true,
// this will parse as much input as possible. Otherwise, it will take only as
// much input as needed to make a valid formula.
func (parser *parser) expectSaysOrSpeaksfor(greedy bool) (Form, error) {
	// Prin [from Time] [until Time] says Form
	// Prin speaksfor Prin
	p, err := parser.parsePrin()
	if err != nil {
		return nil, err
	}
	switch parser.cur() {
	case tokenSpeaksfor:
		parser.advance()
		d, err := parser.parsePrin()
		if err != nil {
			return nil, err
		}
		return Speaksfor{p, d}, nil
	case tokenFrom, tokenUntil, tokenSays:
		from, err := parser.expectOptionalTime(tokenFrom)
		if err != nil {
			return nil, err
		}
		until, err := parser.expectOptionalTime(tokenUntil)
		if err != nil {
			return nil, err
		}
		if from == nil {
			from, err = parser.expectOptionalTime(tokenFrom)
			if err != nil {
				return nil, err
			}
		}
		if parser.cur() != tokenSays {
			if from == nil && until == nil {
				return nil, fmt.Errorf(`expected "from", "until" or "says", found %v`, parser.cur())
			} else if until == nil {
				return nil, fmt.Errorf(`expected "until" or "says", found %v`, parser.cur())
			} else if from == nil {
				return nil, fmt.Errorf(`expected "from" or "says", found %v`, parser.cur())
			} else {
				return nil, fmt.Errorf(`expected "says", found %v`, parser.cur())
			}
		}
		parser.advance()
		var msg Form
		if greedy {
			msg, err = parser.parseForm()
		} else {
			msg, err = parser.parseFormAtHigh(true)
		}
		if err != nil {
			return nil, err
		}
		return Says{p, from, until, msg}, nil
	default:
		return nil, fmt.Errorf(`expected "speaksfor", "from", "until", or "says", found %v`, parser.cur())
	}
}

// The functions follow normal precedence rules, e.g. roughly:
// L = O imp I | I
// O = A or A or A or ... or A | A
// A = H and H and H ... and H | H
// H = not N | ( L ) | P(x) | true | false | P says L | P speaksfor P

// parseFormAtHigh parses a Form, but stops at any binary Form operator. If
// greedy is true, this will parse as much input as possible. Otherwise, it will
// parse only as much input as needed to make a valid formula.
func (parser *parser) parseFormAtHigh(greedy bool) (Form, error) {
	switch parser.cur() {
	case tokenLP:
		parser.advance()
		f, err := parser.parseForm()
		if err != nil {
			return nil, err
		}
		err = parser.expect(tokenRP)
		if err != nil {
			return nil, err
		}
		return f, nil
	case tokenTrue, tokenFalse:
		return parser.expectConst()
	case tokenNot:
		parser.advance()
		f, err := parser.parseFormAtHigh(greedy)
		if err != nil {
			return nil, err
		}
		return Not{f}, nil
	case tokenKey, tokenTPM:
		return parser.expectSaysOrSpeaksfor(greedy)
	default:
		if parser.cur().typ == itemIdentifier {
			return parser.expectPred()
		}
		return nil, fmt.Errorf("expected Form, found %v", parser.cur())
	}
}

// parseFormAtAnd parses a Form, but stops when it reaches a binary Form
// operator of lower precedence than "and".
func (parser *parser) parseFormAtAnd() (Form, error) {
	f, err := parser.parseFormAtHigh(true)
	if err != nil {
		return nil, err
	}
	if parser.cur() != tokenAnd {
		return f, nil
	}
	and, ok := f.(And)
	if !ok {
		and = And{Conjunct: []Form{f}}
	}
	for parser.cur() == tokenAnd {
		parser.advance()
		g, err := parser.parseFormAtHigh(true)
		if err != nil {
			return nil, err
		}
		and.Conjunct = append(and.Conjunct, g)
	}
	return and, nil
}

// parseFormAtOr parses a Form, but stops when it reaches a binary Form operator
// of lower precedence than "or".
func (parser *parser) parseFormAtOr() (Form, error) {
	f, err := parser.parseFormAtAnd()
	if err != nil {
		return nil, err
	}
	if parser.cur() != tokenOr {
		return f, nil
	}
	or, ok := f.(Or)
	if !ok {
		or = Or{Disjunct: []Form{f}}
	}
	for parser.cur() == tokenOr {
		parser.advance()
		g, err := parser.parseFormAtAnd()
		if err != nil {
			return nil, err
		}
		or.Disjunct = append(or.Disjunct, g)
	}
	return or, nil
}

// parseForm parses a Form. This function is greedy: it consumes as much input
// as possible until either an error or EOF is encountered.
func (parser *parser) parseForm() (Form, error) {
	f, err := parser.parseFormAtOr()
	if err != nil {
		return nil, err
	}
	if parser.cur() != tokenImplies {
		return f, nil
	}
	parser.advance()
	g, err := parser.parseForm()
	if err != nil {
		return nil, err
	}
	return Implies{f, g}, nil
}

// parseShortestForm parses the shortest valid Form. This function is not
// greedy: it consumes only as much input as necessary to obtain a valid
// formula. For example, "(p says a and b ...)" and "p says (a and b ...) will
// be parsed in their entirety, but given "p says a and b ... ", only "p says a"
// will be parsed.
func (parser *parser) parseShortestForm() (Form, error) {
	return parser.parseFormAtHigh(false)
}

func newParser(input reader) *parser {
	lex := lex(input)
	return &parser{lex: lex}
}