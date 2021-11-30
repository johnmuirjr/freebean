/*
Copyright (c) 2021, Jordan Vaughan
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package parser

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode"
)

var (
	escapingAtEofError error = errors.New("unfinished escape at end of file")
	inStringAtEofError error = errors.New("unfinished quoted string at end of file")
)

// TokenType is an enum representing different types of lexed tokens.
type TokenType int

const (
	// String indicates an unquoted string.
	String TokenType = iota

	// QuotedString indicates a quoted string.
	QuotedString

	// OpenParen indicates an opening parenthesis ('(').
	OpenParen

	// CloseParen indicates an closing parenthesis (')').
	CloseParen

	// Error represents a syntax error or io.EOF.
	Error

	// none is an internal TokenType indicating that no token has been
	// lexed yet.
	none
)

// Lexer is a simple token lexer.
type Lexer struct {
	reader           *bufio.Reader
	lineNumber       uint64
	isEscaping       bool
	isInString       bool
	isInQuotedString bool // only meaningful when isInString
	token            strings.Builder
	openParenSet     bool
	closeParenSet    bool
}

// NewLexer constructs a Lexer for the specified io.Reader.
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		reader:     bufio.NewReader(r),
		lineNumber: 1}
}

// Get the Lexer's current line number.
func (l *Lexer) LineNumber() uint64 {
	return l.lineNumber
}

// GetNextToken lexes the next token from the Lexer's io.Reader.
// The returned error is io.EOF if the Lexer reached the end of the io.Reader.
// If the returned TokenType is Error, then the returned error is either
// a syntax error or io.EOF.  Note that GetNextToken may return io.EOF
// even when the TokenType is not Error.  The returned string is valid only
// when th TokenType is either String or QuotedString.
func (l *Lexer) GetNextToken() (TokenType, string, error) {
	if l.openParenSet {
		l.openParenSet = false
		return OpenParen, "", nil
	} else if l.closeParenSet {
		l.closeParenSet = false
		return CloseParen, "", nil
	}
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return l.getFinalToken()
			}
			return Error, "", err
		}
		tokenType, token := l.addRuneAndGetToken(r)
		if tokenType == OpenParen || tokenType == CloseParen {
			return tokenType, "", nil
		} else if tokenType != none {
			return tokenType, token, nil
		}
	}
}

// addRuneAndGetToken processes the specified rune and returns a token, if any.
func (l *Lexer) addRuneAndGetToken(r rune) (tokenType TokenType, token string) {
	tokenType = none
	token = ""
	isNewline := r == '\n'
	isSpace := unicode.IsSpace(r)
	if isNewline {
		l.lineNumber++
	}

	if l.isEscaping {
		l.token.WriteRune(r)
		l.isEscaping = false
		if !l.isInString {
			l.isInString = true
		}
	} else if r == '\\' {
		l.isEscaping = true
	} else if l.isInQuotedString {
		if r == '"' {
			token = l.token.String()
			l.token.Reset()
			l.isInString = false
			l.isInQuotedString = false
			tokenType = QuotedString
		} else {
			l.token.WriteRune(r)
		}
	} else if l.isInString {
		if r == '"' {
			token = l.token.String()
			l.token.Reset()
			l.isInQuotedString = true
			tokenType = String
		} else if r == '(' {
			token = l.token.String()
			l.token.Reset()
			l.isInString = false
			l.openParenSet = true
			tokenType = String
		} else if r == ')' {
			token = l.token.String()
			l.token.Reset()
			l.isInString = false
			l.closeParenSet = true
			tokenType = String
		} else if isSpace {
			token = l.token.String()
			l.token.Reset()
			l.isInString = false
			tokenType = String
		} else {
			l.token.WriteRune(r)
		}
	} else if isSpace {
		// do nothing
	} else if r == '"' {
		l.isInString = true
		l.isInQuotedString = true
	} else if r == '(' {
		tokenType = OpenParen
	} else if r == ')' {
		tokenType = CloseParen
	} else {
		l.token.WriteRune(r)
		l.isInString = true
	}
	return
}

// getFinalToken returns the stream's final token or an error if the Lexer
// is in an invalid state at EOF.  This should be called only when the
// Lexer reaches its io.Reader's EOF.
func (l *Lexer) getFinalToken() (tokenType TokenType, token string, e error) {
	tokenType = Error
	if l.isInQuotedString {
		e = inStringAtEofError
	} else if l.isEscaping {
		e = escapingAtEofError
	} else if !l.isInString {
		e = io.EOF
	} else {
		tokenType = String
		token = l.token.String()
		l.isInString = false
	}
	return
}
