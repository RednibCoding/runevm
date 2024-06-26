package runevm

import (
	"fmt"
	"strings"
	"unicode"
)

type Token struct {
	Type   string
	Value  string
	File   string
	Line   int
	Col    int
	Length int
}

type TokenStream struct {
	input    *InputStream
	current  *Token
	last     *Token
	keywords map[string]bool
}

func newTokenStream(input *InputStream) *TokenStream {
	keywords := map[string]bool{
		"if": true, "then": true, "elif": true, "else": true, "while": true, "break": true, "continue": true, "fun": true, "return": true,
		"true": true, "false": true, "array": true, "table": true, "import": true, "not": true,
	}
	return &TokenStream{input: input, keywords: keywords}
}

func (ts *TokenStream) isKeyword(x string) bool {
	return ts.keywords[x]
}

func (ts *TokenStream) isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

func (ts *TokenStream) isIdStart(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func (ts *TokenStream) isId(ch byte) bool {
	return ts.isIdStart(ch) || strings.ContainsRune("?!-<>=0123456789", rune(ch))
}

func (ts *TokenStream) isOpChar(ch byte) bool {
	return strings.ContainsRune("+-*/%=&|<>!", rune(ch))
}

func (ts *TokenStream) isPunc(ch byte) bool {
	return strings.ContainsRune(".,:(){}[]", rune(ch))
}

func (ts *TokenStream) isWhitespace(ch byte) bool {
	return strings.ContainsRune(" \r\t\n", rune(ch))
}

func (ts *TokenStream) readWhile(predicate func(byte) bool) (string, int) {
	var str strings.Builder
	startPos := ts.input.Pos
	for !ts.input.eof() && predicate(ts.input.peek()) {
		str.WriteByte(ts.input.next())
	}
	return str.String(), ts.input.Pos - startPos
}

func (ts *TokenStream) readNumber() *Token {
	number, length := ts.readWhile(func(ch byte) bool {
		if ch == '.' {
			return true
		}
		return ts.isDigit(ch)
	})
	return &Token{Type: "num", Value: number, File: ts.input.filepath, Line: ts.input.line, Col: ts.input.Col - length, Length: length}
}

func (ts *TokenStream) readIdent() *Token {
	id, length := ts.readWhile(ts.isId)
	if ts.isKeyword(id) {
		return &Token{Type: "kw", Value: id, File: ts.input.filepath, Line: ts.input.line, Col: ts.input.Col - length, Length: length}
	}
	return &Token{Type: "var", Value: id, File: ts.input.filepath, Line: ts.input.line, Col: ts.input.Col - length, Length: length}
}

func (ts *TokenStream) readEscaped(end byte) (string, int) {
	var escaped bool
	var str strings.Builder
	startPos := ts.input.Pos
	ts.input.next() // Consume initial quote
	for !ts.input.eof() {
		ch := ts.input.next()
		if escaped {
			str.WriteByte(ch)
			escaped = false
		} else if ch == '\\' {
			escaped = true
		} else if ch == end {
			break
		} else {
			str.WriteByte(ch)
		}
	}
	return str.String(), ts.input.Pos - startPos
}

func (ts *TokenStream) readString() *Token {
	str, length := ts.readEscaped('"')
	return &Token{Type: "str", Value: str, File: ts.input.filepath, Line: ts.input.line, Col: ts.input.Col - length, Length: length}
}

func (ts *TokenStream) skipComment() {
	ts.readWhile(func(ch byte) bool { return ch != '\n' })
	ts.input.next()
}

func (ts *TokenStream) readNext() *Token {
	ts.readWhile(ts.isWhitespace)

	if ts.input.eof() {
		return nil
	}
	ch := ts.input.peek()

	switch {
	case ch == '#':
		ts.skipComment()
		return ts.readNext()
	case ch == '"':
		return ts.readString()
	case ts.isDigit(ch):
		return ts.readNumber()
	case ts.isIdStart(ch):
		return ts.readIdent()
	case ts.isPunc(ch):
		length := 1
		return &Token{Type: "punc", Value: string(ts.input.next()), File: ts.input.filepath, Line: ts.input.line, Col: ts.input.Col - length, Length: length}
	case ts.isOpChar(ch):
		op, length := ts.readWhile(ts.isOpChar)
		return &Token{Type: "op", Value: op, File: ts.input.filepath, Line: ts.input.line, Col: ts.input.Col - length, Length: length}
	default:
		errTok := &Token{Type: "", Value: "", File: ts.input.filepath, Line: ts.input.line, Col: ts.input.Col, Length: 0}
		ts.input.error(errTok, fmt.Sprintf("invalid character: %c", ch))
		return nil
	}
}

func (ts *TokenStream) peek() *Token {
	if ts.current == nil {
		ts.current = ts.readNext()
	}
	return ts.current
}

func (ts *TokenStream) next() *Token {
	tok := ts.current
	ts.current = nil
	if tok == nil {
		tok = ts.readNext()
	}
	ts.last = ts.copyToken(tok)
	return tok
}

func (ts *TokenStream) eof() bool {
	return ts.peek() == nil
}

func (ts *TokenStream) error(tok *Token, msg string) {
	ts.input.error(tok, msg)
}

func (ts *TokenStream) copyToken(tok *Token) *Token {
	t := &Token{
		Type:   tok.Type,
		Value:  tok.Value,
		File:   tok.File,
		Line:   tok.Line,
		Col:    tok.Col,
		Length: tok.Length,
	}
	return t
}
