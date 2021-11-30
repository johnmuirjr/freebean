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
	"reflect"
	"strings"
	"testing"
)

func TestOperands_Length(t *testing.T) {
	values := []interface{}{1, 2, 3}
	for n := 0; n < len(values); n++ {
		op := Operands{stack: &values, stackIndex: n}
		if op.Length() != len(values)-n {
			t.Errorf("expected Operands with %v values and stack index %v to have length %v, but length is %v", len(values), n, len(values)-n, op.Length())
		}
	}
}

func TestOperands_GetValues(t *testing.T) {
	values := []interface{}{1, 2, 3}
	for n := 0; n < len(values); n++ {
		op := Operands{stack: &values, stackIndex: n}
		expected := values[n:]
		if !reflect.DeepEqual(op.GetValues(), expected) {
			t.Errorf("GetValues() with stack index %v returned unexpected slice: %v", n, op.GetValues())
		}
	}
}

func TestOperands_Push(t *testing.T) {
	values := []interface{}{1, 2, 3}
	op := Operands{stack: &values}
	op.Push(4, 5)
	if !reflect.DeepEqual(op.GetValues(), []interface{}{1, 2, 3, 4, 5}) {
		t.Errorf("Push() failed: GetValues() doesn't return the old and new values")
	}
	if !reflect.DeepEqual(values, op.GetValues()) {
		t.Errorf("Push() failed: stack is unmodified")
	}
}

func TestOperands_Pop(t *testing.T) {
	values := []interface{}{1, 2, 3, 4, 5}
	op := Operands{stack: &values}
	popped := op.Pop(2)
	if !reflect.DeepEqual(op.GetValues(), []interface{}{1, 2, 3}) {
		t.Errorf("Pop() failed: GetValues() doesn't return the old and new values")
	}
	if !reflect.DeepEqual(values, []interface{}{1, 2, 3}) {
		t.Errorf("Pop() failed: stack is unmodified")
	}
	if !reflect.DeepEqual(popped, []interface{}{4, 5}) {
		t.Errorf("Pop() didn't return the popped values, got %v", popped)
	}
}

func TestOperands_Pop_TooManyValues(t *testing.T) {
	values := []interface{}{1, 2, 3, 4, 5}
	op := Operands{stack: &values, stackIndex: 3}
	popped := op.Pop(5)
	if len(op.GetValues()) != 0 {
		t.Errorf("Pop() failed: GetValues() doesn't return an empty slice: %v", op.GetValues())
	}
	if !reflect.DeepEqual(values, []interface{}{1, 2, 3}) {
		t.Errorf("Pop() failed: stack is different than expected: %v", values)
	}
	if !reflect.DeepEqual(popped, []interface{}{4, 5}) {
		t.Errorf("Pop() didn't return the popped values, got %v", popped)
	}
}

func TestParser_Parse_EmptyInputNoFunctions(t *testing.T) {
	lex := NewLexer(strings.NewReader(""))
	p := NewParser(nil)
	if e := p.Parse(lex); e != nil {
		t.Errorf("Parse returned a non-nil error: %v", e)
	}
}

func TestParser_Parse_TokensNoFunctions(t *testing.T) {
	lex := NewLexer(strings.NewReader("token1 token2"))
	p := NewParser(nil)
	if e := p.Parse(lex); e != nil {
		t.Errorf("Parse returned a non-nil error: %v", e)
	}
}

func TestParser_Parse_FunctionCall(t *testing.T) {
	lex := NewLexer(strings.NewReader("token1 token2 test"))
	p := NewParser(t)
	p.Functions["test"] = func(fn string, op Operands, ctx interface{}) error {
		if fn != "test" {
			t.Errorf("test function called but received other function name: %v", fn)
		}
		if op.Length() != 2 {
			t.Errorf("test function received %v operands instead of 2", op.Length())
		} else {
			values := op.Pop(2)
			if values[0].(string) != "token1" {
				t.Errorf("first token is %v instead of token1", values[0])
			} else if values[1].(string) != "token2" {
				t.Errorf("second token is %v instead of token2", values[1])
			}
		}
		if ctx != t {
			t.Errorf("test function given context other than t.Testing object")
		}
		return nil
	}
	if e := p.Parse(lex); e != nil {
		t.Errorf("Parse returned a non-nil error: %v", e)
	}
}

func TestParser_Parse_FunctionCallInsideParentheses(t *testing.T) {
	lex := NewLexer(strings.NewReader("token2 (token2 token3 test) token3 test"))
	p := NewParser(t)
	p.Functions["test"] = func(fn string, op Operands, ctx interface{}) error {
		if op.Length() != 2 {
			t.Errorf("test function received %v operands instead of 2", op.Length())
		} else {
			values := op.Pop(2)
			if values[0].(string) != "token2" {
				t.Errorf("first token is %v instead of token1", values[0])
			} else if values[1].(string) != "token3" {
				t.Errorf("second token is %v instead of token2", values[1])
			}
		}
		return nil
	}
	if e := p.Parse(lex); e != nil {
		t.Errorf("Parse returned a non-nil error: %v", e)
	}
}

func TestParser_Parse_FunctionErrorPassesThrough(t *testing.T) {
	lex := NewLexer(strings.NewReader("token1 token2 error"))
	p := NewParser(t)
	err := fmt.Errorf("error")
	p.Functions["error"] = func(fn string, op Operands, ctx interface{}) error {
		return err
	}
	if e := p.Parse(lex); e.Error() != fmt.Sprintf(`1: %v`, err) {
		t.Errorf("Parse returned unexpected error: %v", e)
	}
}

func TestParser_Parse_QuotedStringsAndParentheses(t *testing.T) {
	lex := NewLexer(strings.NewReader(`"token1"("token2""token3" popall)"token4"`))
	p := NewParser(nil)
	p.Functions["popall"] = func(fn string, op Operands, ctx interface{}) error {
		op.Pop(op.Length())
		return nil
	}
	if e := p.Parse(lex); e != nil {
		t.Errorf("Parse returned a non-nil error: %v", e)
	}
}

func TestParser_Finish_EmptyInput(t *testing.T) {
	lex := NewLexer(strings.NewReader(""))
	p := NewParser(nil)
	p.Parse(lex)
	if e := p.Finish(); e != nil {
		t.Errorf("Finish returned a non-nil error: %v", e)
	}
}

func TestParser_Finish_UnclosedParentheses(t *testing.T) {
	lex := NewLexer(strings.NewReader("()(()"))
	p := NewParser(nil)
	p.Parse(lex)
	if e := p.Finish(); e == nil {
		t.Errorf("Finish returned a nil error")
	}
}

func TestParser_Finish_UnconsumedOperands(t *testing.T) {
	lex := NewLexer(strings.NewReader("token1 token2"))
	p := NewParser(nil)
	p.Parse(lex)
	if e := p.Finish(); e == nil {
		t.Errorf("Finish returned a nil error")
	}
}

func TestSilence(t *testing.T) {
	lex := NewLexer(strings.NewReader(`(silence fail)`))
	p := NewParser(nil)
	p.Functions["fail"] = func(fn string, op Operands, ctx interface{}) error {
		return fmt.Errorf("test failed")
	}
	if e := p.Parse(lex); e != nil {
		t.Errorf("Parse returned a non-nil error: %v", e)
	}
}

func TestSilence_OutsideParens(t *testing.T) {
	lex := NewLexer(strings.NewReader(`silence`))
	p := NewParser(nil)
	if p.Parse(lex) == nil {
		t.Errorf("Parse succeeded but should have failed")
	}
}

func TestSilence_ClosingParenDisablesSilence(t *testing.T) {
	lex := NewLexer(strings.NewReader(`(silence inc) inc`))
	p := NewParser(nil)
	value := 0
	p.Functions["inc"] = func(fn string, op Operands, ctx interface{}) error {
		value++
		return nil
	}
	if err := p.Parse(lex); err != nil {
		t.Errorf("Parse failed: %v", err)
	} else if value != 1 {
		t.Errorf("silence did not silence function execution")
	}
}

func TestSilence_SilenceSilencesNestedParens(t *testing.T) {
	lex := NewLexer(strings.NewReader(`(inc silence (inc silence inc) inc) inc`))
	p := NewParser(nil)
	value := 0
	p.Functions["inc"] = func(fn string, op Operands, ctx interface{}) error {
		value++
		return nil
	}
	if err := p.Parse(lex); err != nil {
		t.Errorf("Parse failed: %v", err)
	} else if value != 2 {
		t.Errorf("silence did not silence function execution inside nested parens")
	}
}

func TestSilence_SilenceInsideNestedParens(t *testing.T) {
	lex := NewLexer(strings.NewReader(`(inc (silence inc) inc) inc`))
	p := NewParser(nil)
	value := 0
	p.Functions["inc"] = func(fn string, op Operands, ctx interface{}) error {
		value++
		return nil
	}
	if err := p.Parse(lex); err != nil {
		t.Errorf("Parse failed: %v", err)
	} else if value != 3 {
		t.Errorf("silence did not stop with nested parens")
	}
}

func TestSilence_AtTopLevelBetweenParens(t *testing.T) {
	lex := NewLexer(strings.NewReader(`(inc) silence inc (inc) inc`))
	p := NewParser(nil)
	value := 0
	p.Functions["inc"] = func(fn string, op Operands, ctx interface{}) error {
		value++
		return nil
	}
	if p.Parse(lex) == nil {
		t.Errorf("Parse succeeded but should have failed")
	}
}
