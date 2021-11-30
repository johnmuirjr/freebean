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
	"io"
	"strings"
	"testing"
)

type token struct {
	tokenType TokenType
	text      string
}

func checkLexer(t *testing.T, input string, tokens []token) {
	lex := NewLexer(strings.NewReader(input))
	for index, expectedToken := range tokens {
		tokenType, text, e := lex.GetNextToken()
		if tokenType != expectedToken.tokenType {
			t.Errorf("expected token %v to be type %v but got type %v", index, expectedToken.tokenType, tokenType)
		} else if tokenType == String && text != expectedToken.text {
			t.Errorf("expected token %v to be string \"%v\" but got \"%v\"", index, expectedToken.text, text)
		}

		if e == io.EOF {
			t.Errorf("unexpected EOF at token %v", index)
		} else if e != nil {
			t.Errorf("got error at token %v: %v", index, e)
		}

		if t.Failed() {
			t.FailNow()
		}
	}
	tokenType, text, e := lex.GetNextToken()
	if tokenType != Error || text != "" || e != io.EOF {
		t.Errorf("unexpected token type %v, text \"%v\", and error \"%v\" after %v tokens", tokenType, text, e, len(tokens))
		t.FailNow()
	}
}

func TestGetNextToken_EmptyInput(t *testing.T) {
	lex := NewLexer(strings.NewReader(""))
	tokenType, text, e := lex.GetNextToken()
	if tokenType != Error {
		t.Errorf("empty input returned token type %v instead of Error", tokenType)
	}
	if text != "" {
		t.Errorf("empty input returned nonempty token: %v", text)
	}
	if e != io.EOF {
		t.Errorf("empty input returned error other than io.EOF: %v", e)
	}
}

func TestGetNextToken_OneString(t *testing.T) {
	checkLexer(t, "someText", []token{{String, "someText"}})
	checkLexer(t, "\t someText\t ", []token{{String, "someText"}})
}

func TestGetNextToken_TwoStrings(t *testing.T) {
	checkLexer(t, "token1 token2", []token{{String, "token1"}, {String, "token2"}})
	checkLexer(t, "token1\ttoken2", []token{{String, "token1"}, {String, "token2"}})
	checkLexer(t, "token1\vtoken2", []token{{String, "token1"}, {String, "token2"}})
	checkLexer(t, "token1\rtoken2", []token{{String, "token1"}, {String, "token2"}})
	checkLexer(t, "token1\ntoken2", []token{{String, "token1"}, {String, "token2"}})
}

func TestGetNextToken_OnlyParens(t *testing.T) {
	checkLexer(t, "() ) (", []token{{OpenParen, ""}, {CloseParen, ""}, {CloseParen, ""}, {OpenParen, ""}})
}

func TestGetNextToken_TokensWithinParens(t *testing.T) {
	checkLexer(t, "(token1) token2( token3 )", []token{{OpenParen, ""}, {String, "token1"}, {CloseParen, ""}, {String, "token2"}, {OpenParen, ""}, {String, "token3"}, {CloseParen, ""}})
}

func TestGetNextToken_QuotedAndUnquotedStrings(t *testing.T) {
	checkLexer(t, "unq1 \"q 1\"", []token{{String, "unq1"}, {QuotedString, "q 1"}})
}

func TestGetNextToken_QuotesTerminateStrings(t *testing.T) {
	checkLexer(t, "unq1\"q 1\"unq2\"q 2\"\"q 3\"", []token{{String, "unq1"}, {QuotedString, "q 1"}, {String, "unq2"}, {QuotedString, "q 2"}, {QuotedString, "q 3"}})
}
