// Copyright 2018 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package properties

import (
	"fmt"
	"runtime"
	"strings"
)

type parser struct {
	lex *lexer
}

func parse(input string, preserveFormatting bool) (properties *Properties, err error) {
	l := lex(input)
	l.keepWS = preserveFormatting
	go l.run()
	p := &parser{lex: l}
	defer p.recover(&err)

	properties = NewProperties()
	key := ""
	comments := []prefixedComment{}

	for {
		token := p.expectOneOf(itemComment, itemKey, itemEOF)
		switch token.typ {
		case itemEOF:
			if !preserveFormatting || (len(comments) == 0 && token.val == "") {
				goto done
			}
			// There are comments at the end of the input that are not tied to a particular key
			// Save these off when preserving formatting
			if token.val == "" {
				// Save off previously parsed trailing comments
				properties.trailingComments = comments
				goto done
			}
			// Need to save off the last comment line as well
			prefixIndex := 0
			// Include leading whitespace into the prefix
			prefixIndex = strings.Index(token.val, strings.TrimSpace(token.val))
			prefix := token.val[0 : prefixIndex+1]
			comment := prefixedComment{prefix, token.val[prefixIndex+1 : len(token.val)]}
			comments = append(comments, comment)
			properties.trailingComments = comments
			goto done
		case itemComment:
			prefix := "#"
			prefixIndex := 0
			if preserveFormatting {
				// Include leading whitespace into the prefix
				prefixIndex = strings.Index(token.val, strings.TrimSpace(token.val))
				prefix = token.val[0:prefixIndex+1]
			}
			comment := prefixedComment{prefix, token.val[prefixIndex+1:len(token.val)]}
			comments = append(comments, comment)
			continue
		case itemKey:
			key = strings.TrimSpace(token.val)
			if _, ok := properties.m[key]; !ok {
				properties.k = append(properties.k, key)
			}
		}

		token = p.expectOneOf(itemValue, itemEOF)
		if len(comments) > 0 {
			properties.c[key] = comments
			comments = []prefixedComment{}
		}
		switch token.typ {
		case itemEOF:
			properties.m[key] = ""
			goto done
		case itemValue:
			properties.m[key] = token.val
		}
	}

done:
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

func (p *parser) expectOneOf(expected ...itemType) (token item) {
	token = p.lex.nextItem()
	for _, v := range expected {
		if token.typ == v {
			return token
		}
	}
	p.unexpected(token)
	panic("unexpected token")
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
