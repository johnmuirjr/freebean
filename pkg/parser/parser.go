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
	"fmt"
	"io"
)

// Function is a custom function that can be registered with a Parser.
// Parsed code can call the function by specifying the function's unquoted name.
// The Parser passes the Function's name (as registered with the Parser),
// the Operands for the function call, and the Parser's context to the Function.
type Function func(string, Operands, interface{}) error

// Parser treats a stream of lexed tokens as a reverse Polish notation language.
// It maintains two stacks: an operand stack containing arbitrary values
// and a "marker stack" of indices into the operand stack.  The top value in
// the marker stack determines where a called Function's operands begin -- that
// is, how far into the stack a called Function can see.  Lexed parentheses
// modify the marker stack (open parentheses push markers, close parentheses
// pop them).
//
// Clients can add arbitrary Functions via the Functions field.  A lexed
// unquoted String calls the Function of the same name; if no such function
// exists, Parser pushes the String onto the operand stack.  QuotedString
// tokens never call functions.
//
// Parser always provides one special function, "silence", that disables
// pushing operands and executing functions until the current marker is popped
// (that is, until the opening parenthesis that precedes the "silence"
// is closed).  This provides a handy way to turn entire blocks of code
// into comments or to disable them for debugging without having to turn
// them into comment strings.  "silence" MUST appear within a pair
// of parentheses: Parsers return errors when they encounter "silence"
// outside of parentheses.
//
// Clients can give Parsers arbitrary context values.  Parser passes the context
// objects to Functions; this allows the latter to maintain state.
type Parser struct {
	operandStack []interface{}
	markerStack  []int
	silenced     int

	// Functions is a case-senstitive registry of Functions.
	Functions map[string]Function

	// Context is an arbitrary value that Parser will pass to
	// called Functions.
	Context interface{}
}

// NewParser creates a new Parser with the specified context.
// The Parser will have empty operand and marker stacks and will have
// no Functions.
func NewParser(context interface{}) *Parser {
	return &Parser{operandStack: make([]interface{}, 0), markerStack: make([]int, 0), Functions: make(map[string]Function), Context: context}
}

func (p *Parser) formatError(lex *Lexer, err error) error {
	return fmt.Errorf(`%v: %v`, lex.LineNumber(), err)
}

// Parse executes the stream of tokens from the specified Lexer.
// It returns nil when the Lexer reaches EOF without problems.
// If a called Function returns an error, Parse stops and returns it unmodified.
func (p *Parser) Parse(lex *Lexer) error {
	for {
		tokenType, text, e := lex.GetNextToken()
		switch tokenType {
		case String:
			if p.silenced == 0 {
				if text == "silence" {
					if len(p.markerStack) == 0 {
						return p.formatError(lex, fmt.Errorf(`found "silence" outside parentheses`))
					}
					p.silenced = len(p.markerStack)
				} else if f, ok := p.Functions[text]; ok {
					if e = f(text, p.getOperands(), p.Context); e != nil {
						return p.formatError(lex, e)
					}
				} else {
					p.pushString(text)
				}
			}
		case QuotedString:
			if p.silenced == 0 {
				p.pushString(text)
			}
		case OpenParen:
			p.markerStack = append(p.markerStack, len(p.operandStack))
		case CloseParen:
			if e = p.onCloseParen(); e != nil {
				return p.formatError(lex, e)
			}
		case Error:
			if e == io.EOF {
				return nil
			}
			return p.formatError(lex, fmt.Errorf(`syntax error: %v`, e))
		default:
			panic("unexpected TokenType")
		}

		if e == io.EOF {
			return nil
		}
	}
}

// Finish runs final checks on the operand and marker stacks.
// It returns nil if there are no problems.
func (p *Parser) Finish() error {
	if len(p.operandStack) > 0 {
		return fmt.Errorf("%v unconsumed tokens left on stack at EOF", len(p.operandStack))
	} else if len(p.markerStack) > 0 {
		return fmt.Errorf("%v unclosed parentheses at EOF", len(p.markerStack))
	} else if p.silenced != 0 {
		return fmt.Errorf("parser evaluation silenced at EOF")
	}
	return nil
}

// pushString is a convenience function for pushing a string onto
// the operand stack.
func (p *Parser) pushString(text string) {
	p.operandStack = append(p.operandStack, text)
}

// getOperands constructs an Operands object using the marker stack's top value.
func (p *Parser) getOperands() Operands {
	index := 0
	if len(p.markerStack) != 0 {
		index = p.markerStack[len(p.markerStack)-1]
		if index > len(p.operandStack) {
			panic("top of marker stack extends beyond length of operand stack")
		}
	}
	return Operands{stack: &p.operandStack, stackIndex: index}
}

// onCloseParen implements the close parenthesis behavior.  It checks whether
// all operand stack values since the last open parenthesis have been popped.
func (p *Parser) onCloseParen() error {
	if len(p.markerStack) == 0 {
		return fmt.Errorf("closing parenthesis does not have a matching open parenthesis")
	} else if len(p.markerStack) == p.silenced {
		p.silenced = 0
	}
	index := p.markerStack[len(p.markerStack)-1]
	p.markerStack = p.markerStack[0 : len(p.markerStack)-1]
	if index != len(p.operandStack) {
		return fmt.Errorf("%v unconsumed operands at closing parenthesis", len(p.operandStack)-index)
	}
	return nil
}
