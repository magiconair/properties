// Copyright 2013 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goproperties

import (
	"fmt"
	"runtime"
)

type parser struct {
	lex *lexer
}

func parse(input string) (properties *Properties, err error) {
	p := &parser{lex: lex(input)}
	defer p.recover(&err)

	p.lex = lex(input)
	properties = NewProperties()

	for {
		token := p.expectOneOf(itemKey, itemEOF)
		if token.typ == itemEOF {
			break
		}
		key := token.val

		token = p.expectOneOf(itemValue, itemEOF)
		if token.typ == itemEOF {
			properties.m[key] = ""
			break
		}
		properties.m[key] = token.val
	}

	return properties, nil
}

func (p *parser) errorf(format string, args ...interface{}) {
	format = fmt.Sprintf("properties: Line %d: %s", p.lex.lineNumber(), format)
	panic(fmt.Errorf(format, args...))
}

func (p *parser) expect(expected itemType) (token item) {
	token = p.lex.nextItem()
	if token.typ != expected {
		p.unexpected(token)
	}
	return token
}

func (p *parser) expectOneOf(expected1, expected2 itemType) (token item) {
	token = p.lex.nextItem()
	if token.typ != expected1 && token.typ != expected2 {
		p.unexpected(token)
	}
	return token
}

func (p *parser) unexpected(token item) {
	p.errorf(token.String())
}

// recover is the handler that turns panics into returns from the top level of Parse.
func (p *parser) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		*errp = e.(error)
	}
	return
}
