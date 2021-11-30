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
)

type Transaction struct {
	Entity      string
	Description string
	Transfers   []*Transfer
	Notes       map[string]string
}

func getTransferAndNoteOperandStartIndices(op parser.Operands) (transferStartIndex, noteStartIndex int) {
	values := op.GetValues()
	for noteStartIndex = len(values) - 1; noteStartIndex >= 0; noteStartIndex-- {
		if _, ok := values[noteStartIndex].(string); !ok {
			noteStartIndex++
			break
		}
	}
	for transferStartIndex = noteStartIndex - 1; transferStartIndex >= 0; transferStartIndex-- {
		if _, ok := values[transferStartIndex].(*Transfer); !ok {
			transferStartIndex++
			break
		}
	}
	return
}

func checkTransfers(transfers []*Transfer) error {
	q := transfers[0].GetTransferQuantity()
	for _, t := range transfers[1:] {
		tq := t.GetTransferQuantity()
		if tq.Commodity != q.Commodity {
			return fmt.Errorf("transfer to %v uses commodity %v but transfer to %v uses %v", t.Account.Name, tq.Commodity, transfers[0].Account.Name, q.Commodity)
		}
		q.Amount = q.Amount.Add(tq.Amount)
	}
	if !q.Amount.IsZero() {
		return fmt.Errorf("transfers sum to %v, not zero", q)
	}
	return nil
}

// Syntax: ENTITY DESCRIPTION Transfer+ (NOTE-NAME NOTE-VALUE)* xact ->
func ParseTransaction(op parser.Operands, ctx *core.Context) (Transaction, error) {
	t := Transaction{}
	var ok bool
	values := op.GetValues()
	transferStartIndex, noteStartIndex := getTransferAndNoteOperandStartIndices(op)
	if transferStartIndex == 0 {
		return t, fmt.Errorf("entity and description operands are required")
	} else if transferStartIndex == 1 {
		return t, fmt.Errorf("description operand is required")
	}
	numTransfers := noteStartIndex - transferStartIndex
	if numTransfers < 2 {
		return t, fmt.Errorf("there must be at least two transfers")
	}
	numNotes := len(values) - noteStartIndex
	if numNotes%2 != 0 {
		return t, fmt.Errorf("the number of notes must be a multiple of two, got %v", numNotes)
	}
	values = op.Pop(numTransfers + numNotes + 2)
	if t.Entity, ok = values[0].(string); !ok {
		return t, fmt.Errorf("non-string entity: %v", values[0])
	} else if t.Description, ok = values[1].(string); !ok {
		return t, fmt.Errorf("non-string description: %v", values[1])
	}
	t.Transfers = make([]*Transfer, numTransfers)[:0]
	for _, transfer := range values[2 : numTransfers+2] {
		t.Transfers = append(t.Transfers, transfer.(*Transfer))
	}
	if err := checkTransfers(t.Transfers); err != nil {
		return t, err
	}
	t.Notes = make(map[string]string, numNotes)
	for n := numTransfers + 2; n < len(values); n += 2 {
		t.Notes[values[n].(string)] = values[n+1].(string)
	}
	return t, nil
}

func (t *Transaction) Execute(ctx *core.Context) error {
	for _, transfer := range t.Transfers {
		if err := transfer.ExecuteTransfer(ctx); err != nil {
			return err
		}
	}
	return nil
}
