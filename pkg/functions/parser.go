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

package functions

import (
	"fmt"
	"github.com/jtvaughan/freebean/pkg/core"
	"github.com/jtvaughan/freebean/pkg/parser"
	"io"
)

type Function func(string, parser.Operands, *core.Context) error

type Parser struct {
	Functions map[string]Function

	ctx    *core.Context
	lexer  *parser.Lexer
	parser *parser.Parser
}

func NewParser(r io.Reader) *Parser {
	ctx := core.NewContext()
	return &Parser{
		Functions: make(map[string]Function),
		ctx:       ctx,
		lexer:     parser.NewLexer(r),
		parser:    parser.NewParser(ctx)}
}

func (p *Parser) Context() *core.Context { return p.ctx }

func (p *Parser) AddCoreFunctions() {
	for fn, f := range GetCoreFunctions() {
		p.Functions[fn] = f
	}
}

func (p *Parser) Parse() error {
	for fn, f := range p.Functions {
		f := f
		p.parser.Functions[fn] = func(fn string, op parser.Operands, _ interface{}) error {
			return f(fn, op, p.ctx)
		}
	}
	err := p.parser.Parse(p.lexer)
	if err != nil {
		err = fmt.Errorf(`%v: %v`, p.ctx.Date, err)
	} else {
		err = p.parser.Finish()
	}
	return err
}
