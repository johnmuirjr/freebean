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

// Operands is a view of a Parser's operand stack.
// Parsers pass Operands to Functions.  Functions use Operands to view
// and modify the stack, as necessary.  Operands guarantees that Functions
// cannot modify the operand stack outside of the current pair of parentheses.
type Operands struct {
	// pointer so that Push() and Pop() can modify the original stack
	stack *[]interface{}

	// where the operands start in stack
	stackIndex int
}

// GetValues returns all of the Operands values.
func (op *Operands) GetValues() []interface{} {
	return (*op.stack)[op.stackIndex:]
}

// Length returns the number of Operands values.
// This is slightly more efficient than calling len(GetValues()).
func (op *Operands) Length() int {
	return len(*op.stack) - op.stackIndex
}

// Push pushes the specified values onto the associated Parser's operand stack.
// GetValues and Length will include the new values.
func (op *Operands) Push(values ...interface{}) {
	*op.stack = append(*op.stack, values...)
}

// Pop pops the specified number of values from the associated Parser's
// operand stack and returns them.  Pop will not pop more than Length values.
func (op *Operands) Pop(numValues int) []interface{} {
	length := op.Length()
	if numValues > length {
		numValues = length
	}
	stackIndex := len(*op.stack) - numValues
	values := (*op.stack)[stackIndex:]
	*op.stack = (*op.stack)[0:stackIndex]
	return values
}
