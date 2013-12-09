// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package properties

import (
	"fmt"
	// "log"
	"strings"
	// "unicode"
	"strconv"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type Pos int

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is text of error
	itemEOF
	itemDelim // a = or : delimiter char
	itemKey   // a key
	itemValue // a value
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	input   string    // the string being scanned
	state   stateFn   // the next lexing function to enter
	pos     Pos       // current position in the input
	start   Pos       // start position of this item
	width   Pos       // width of last rune read from input
	lastPos Pos       // position of most recent item returned by nextItem
	items   chan item // channel of scanned items
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.emitWithValue(t, l.input[l.start:l.pos])
}

// emitWithValue passes an item with a specific value back to the client.
func (l *lexer) emitWithValue(t itemType, value string) {
	item := item{t, l.start, value}
	// log.Printf("lex.emit: %s", item)
	l.items <- item
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// accept until consumes runes until a termination rune.
func (l *lexer) acceptUntil(term rune) {
	for r := l.next(); r != eof && r != term; {
	}
}

// hasText returns true if the current parsed text is not empty.
func (l *lexer) isNotEmpty() bool {
	return l.pos > l.start
}

// lineNumber reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexKey(l); l.state != nil; {
		l.state = l.state(l)
	}
}

// state functions
// TODO: handle comments
// TODO: handle multi-line values
// TODO: handle unicode literals

// lexKey scans the key up to a delimiter
func lexKey(l *lexer) stateFn {
	if l.peek() == eof {
		l.emit(itemEOF)
		return nil
	}

	runes := make([]rune, 0, 32)

Loop:
	for {
		switch r := l.next(); {

		case r == '\\':
			switch r = l.next(); {

			// escaped key termination chars
			case r == ' ' || r == ':' || r == '=':
				runes = append(runes, r)

			// unicode literals
			case r == 'u' || r == 'U':
				r, err := scanUnicodeLiteral(l)
				if err != nil {
					return l.errorf(err.Error())
				}
				runes = append(runes, r)

			// EOF
			case r == eof:
				return l.errorf("premature EOF")

			// everything else is an error
			default:
				return l.errorf("invalid escape sequence %s", string(r))
			}

		// terminate the key (same as escapes above)
		case r == ' ' || r == ':' || r == '=':
			l.backup()
			break Loop

		case r == eof:
			return l.errorf("premature EOF")

		default:
			runes = append(runes, r)
		}
	}

	if len(runes) > 0 {
		l.emitWithValue(itemKey, string(runes))
	}

	// ignore trailing spaces
	l.acceptRun(" ")
	l.ignore()

	return lexDelim
}

// lexDelim scans the delimiter. We expect to be just before the delimiter
func lexDelim(l *lexer) stateFn {
	if l.next() == eof {
		return l.errorf("premature EOF")
	}
	l.emit(itemDelim)
	return lexValue
}

// lexValue scans text until the end of the line. We expect to be just after the delimiter
func lexValue(l *lexer) stateFn {
	// ignore leading spaces
	l.acceptRun(" ")
	l.ignore()

	runes := make([]rune, 0, 128)
	for {
		switch r := l.next(); {
		// TODO: handle multiline with indent on subsequent lines
		// TODO: handle unicode literals \uXXXX and \Uxxxx
		// TODO: handle escaped chars \n, \r, \t and \\
		case r == '\n':
			l.emitWithValue(itemValue, string(runes))

			// ignore the new line
			l.ignore()
			return lexKey

		case r == eof:
			l.emitWithValue(itemValue, string(runes))
			l.emit(itemEOF)
			return nil

		default:
			runes = append(runes, r)
		}
	}
}

// scans the digits of the unicode literal in \uXXXX form.
// We expect to be before the first digit
func scanUnicodeLiteral(l *lexer) (rune, error) {
	d := make([]rune, 4)
	for i := 0; i < 4; i++ {
		d[i] = l.next()
		if d[i] == eof {
			return eof, nil
		}
	}

	u := string(d)
	s, err := strconv.Unquote(fmt.Sprintf("'\\u%s'", u))
	if err != nil {
		return 0, fmt.Errorf("invalid unicode literal %s", u)
	}

	r, _ := utf8.DecodeRuneInString(s)
	return r, nil
}
