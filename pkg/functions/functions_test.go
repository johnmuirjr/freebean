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
	"github.com/shopspring/decimal"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func createParser(program string) *Parser {
	p := NewParser(strings.NewReader(program))
	p.AddCoreFunctions()
	return p
}

func atoi(fn string, op parser.Operands, ctx *core.Context) error {
	op.Push(strconv.Atoi(op.Pop(1)[0].(string)))
	return nil
}

func TestAddCoreFunctions(t *testing.T) {
	p := NewParser(nil)
	p.AddCoreFunctions()
	for fn, _ := range GetCoreFunctions() {
		if _, ok := p.Functions[fn]; !ok {
			t.Errorf("AddCoreFunctions did not add core function %v", fn)
		}
	}
}

func TestAddNotesFunction(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		Assets:Account open
		Assets:Account type "regular asset" checking yes add-notes)`)
	if e := p.Parse(); e != nil {
		t.Errorf(`add-notes function failed: %v`, e)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf(`open did not create an account`)
	} else if len(a.Notes) != 2 {
		t.Errorf(`add-notes did not add 2 notes, added: %v`, a.Notes)
	} else if n, ok := a.Notes["type"]; !ok {
		t.Errorf(`add-notes did not add a "type" note`)
	} else if n != "regular asset" {
		t.Errorf(`add-notes set "type" note to "%v" instead of "regular asset"`, n)
	} else if n, ok := a.Notes["checking"]; !ok {
		t.Errorf(`add-notes did not add a "checking" note`)
	} else if n != "yes" {
		t.Errorf(`add-notes set "checking" note to "%v" instead of "yes"`, n)
	}
}

func TestAddNotesFunction_ZeroOperands(t *testing.T) {
	p := createParser(`add-notes`)
	if p.Parse() == nil {
		t.Errorf(`add-notes function succeeded but should have failed`)
	}
}

func TestAddNotesFunction_OddNumberOfNoteOperands(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account name add-notes`)
	if p.Parse() == nil {
		t.Errorf(`add-notes function succeeded but should have failed`)
	}
	p = createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account name value name add-notes`)
	if p.Parse() == nil {
		t.Errorf(`add-notes function succeeded but should have failed`)
	}
}

func TestAddNotesFunction_NonStringAccountName(t *testing.T) {
	p := createParser(`123 atoi name value add-notes`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`add-notes function succeeded but should have failed`)
	}
}

func TestAddNotesFunction_NonStringNoteName(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account 123 atoi value add-notes`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`add-notes function succeeded but should have failed`)
	}
}

func TestAddNotesFunction_NonStringNoteValue(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account name 123 atoi add-notes`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`add-notes function succeeded but should have failed`)
	}
}

func TestAddNotesFunction_NonexistentAccount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account type "regular asset" add-notes`)
	if p.Parse() == nil {
		t.Errorf(`add-notes function succeeded but should have failed`)
	}
}

func TestAddNotesFunction_ClosedAccount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account close
		Assets:Account type "regular asset" add-notes`)
	if p.Parse() == nil {
		t.Errorf(`add-notes function succeeded but should have failed`)
	}
}

func TestAddNotesFunction_NoNotes(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		Assets:Account open
		Assets:Account type "regular asset" add-notes
		Assets:Account add-notes)`)
	if e := p.Parse(); e != nil {
		t.Errorf(`add-notes function failed: %v`, e)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf(`open did not create an account`)
	} else if len(a.Notes) != 1 {
		t.Errorf(`add-notes did not add 1 note, added: %v`, a.Notes)
	} else if n, ok := a.Notes["type"]; !ok {
		t.Errorf(`add-notes did not add a "type" note`)
	} else if n != "regular asset" {
		t.Errorf(`add-notes set "type" note to "%v" instead of "regular asset"`, n)
	}
}

func TestAddNotesFunction_DuplicateNotes(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		Assets:Account open
		Assets:Account type "regular asset" type "other" add-notes)`)
	if e := p.Parse(); e != nil {
		t.Errorf(`add-notes function failed: %v`, e)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf(`open did not create an account`)
	} else if len(a.Notes) != 1 {
		t.Errorf(`add-notes did not add 1 note, added: %v`, a.Notes)
	} else if n, ok := a.Notes["type"]; !ok {
		t.Errorf(`add-notes did not add a "type" note`)
	} else if n != "other" {
		t.Errorf(`add-notes set "type" note to "%v" instead of "other"`, n)
	}
}

func TestAssertFunction(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 10 USD xfer
			Equity -10 USD xfer
			xact
		Assets:Account 10 USD assert
		Equity -10 USD assert`)
	if e := p.Parse(); e != nil {
		t.Errorf("assert function failed: %v", e)
	}
}

func TestAssertFunction_WrongAmount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 10 USD xfer
			Equity -10 USD xfer
			xact
		Assets:Account 10.1 USD assert`)
	if p.Parse() == nil {
		t.Errorf("assert function succeeded but should have failed")
	}
}

func TestAssertFunction_NonzeroAmountOfAbsentCommodity(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Assets:Account 1 USD assert`)
	if p.Parse() == nil {
		t.Errorf("assert function succeeded but should have failed")
	}
}

func TestAssertFunction_ZeroAmountOfAbsentCommodity(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Assets:Account 0 USD assert`)
	if e := p.Parse(); e != nil {
		t.Errorf("assert function failed: %v", e)
	}
}

func TestAssertFunction_IgnoresNonDefaultLots(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 1 USD xfer foolot create-lot
			Equity -1 USD xfer
			xact
		Assets:Account 0 USD assert`)
	if e := p.Parse(); e != nil {
		t.Errorf("assert function failed: %v", e)
	}
}

func TestAssertFunction_ZeroOperands(t *testing.T) {
	p := createParser(`assert`)
	if p.Parse() == nil {
		t.Errorf("assert function succeeded but should have failed")
	}
}

func TestAssertFunction_NonStringAccountName(t *testing.T) {
	p := createParser(`
		USD Dollar commodity
		123 atoi 0 USD assert`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("assert function succeeded but should have failed")
	}
}

func TestAssertFunction_IllegalAmount(t *testing.T) {
	p := createParser(`
		USD Dollar commodity
		Assets:Account open
		Assets:Account 0a USD assert`)
	if p.Parse() == nil {
		t.Errorf("assert function succeeded but should have failed")
	}
}

func TestAssertFunction_NonStringCommodityName(t *testing.T) {
	p := createParser(`
		Assets:Account open
		Assets:Account 0 123 atoi assert`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("assert function succeeded but should have failed")
	}
}

func TestAssertFunction_NonexistentAccount(t *testing.T) {
	p := createParser(`
		USD Dollar commodity
		Assets:Account 0 USD assert`)
	if p.Parse() == nil {
		t.Errorf("assert function succeeded but should have failed")
	}
}

func TestAssertFunction_NonexistentCommodity(t *testing.T) {
	p := createParser(`
		Assets:Account open
		Assets:Account 0 USD assert`)
	if p.Parse() == nil {
		t.Errorf("assert function succeeded but should have failed")
	}
}

func TestAssertFunction_ClosedAccount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Assets:Account close
		Assets:Account 0 USD assert`)
	if p.Parse() == nil {
		t.Errorf("assert function succeeded but should have failed")
	}
}

func TestAssertLotFunction(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 10 USD xfer foolot create-lot
			Equity -10 USD xfer barlot create-lot
			xact
		Assets:Account foolot 10 USD assert-lot
		Equity barlot -10 USD assert-lot`)
	if e := p.Parse(); e != nil {
		t.Errorf("assert-lot function failed: %v", e)
	}
}

func TestAssertLotFunction_WrongAmount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 10 USD xfer foolot create-lot
			Equity -10 USD xfer
			xact
		Assets:Account foolot 10.1 USD assert-lot`)
	if p.Parse() == nil {
		t.Errorf("assert-lot function succeeded but should have failed")
	}
}

func TestAssertLotFunction_NonzeroAmountOfAbsentCommodity(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Assets:Account foolot 1 USD assert-lot`)
	if p.Parse() == nil {
		t.Errorf("assert-lot function succeeded but should have failed")
	}
}

func TestAssertLotFunction_ZeroAmountOfAbsentCommodity(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		JPY Yen commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 1 JPY xfer foolot create-lot
			Equity -1 JPY xfer
			xact
		Assets:Account foolot 0 USD assert-lot`)
	if e := p.Parse(); e != nil {
		t.Errorf("assert-lot function failed: %v", e)
	}
}

func TestAssertLotFunction_IgnoresOtherLots(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 1 USD xfer foolot create-lot
			Assets:Account 2 USD xfer barlot create-lot
			Equity -3 USD xfer
			xact
		Assets:Account foolot 1 USD assert-lot
		Assets:Account barlot 2 USD assert-lot`)
	if e := p.Parse(); e != nil {
		t.Errorf("assert-lot function failed: %v", e)
	}
}

func TestAssertLotFunction_TooFewOperands(t *testing.T) {
	for _, program := range []string{"assert-lot", "Assets:Account assert-lot", "Assets:Account foolot assert-lot", "Assets:Account foolot 1 assert-lot"} {
		p := createParser(program)
		if p.Parse() == nil {
			t.Errorf("assert-lot function succeeded but should have failed")
		}
	}
}

func TestAssertLotFunction_NonStringAccountName(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		123 atoi foolot 0 USD assert-lot`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("assert-lot function succeeded but should have failed")
	}
}

func TestAssertLotFunction_IllegalAmount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Assets:Account foolot 0a USD assert-lot`)
	if p.Parse() == nil {
		t.Errorf("assert-lot function succeeded but should have failed")
	}
}

func TestAssertLotFunction_NonStringCommodityName(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account foolot 0 123 atoi assert-lot`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("assert-lot function succeeded but should have failed")
	}
}

func TestAssertLotFunction_NonStringLotName(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Assets:Account 123 atoi 0 USD assert-lot`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("assert-lot function succeeded but should have failed")
	}
}

func TestAssertLotFunction_NonexistentAccount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account foolot 0 USD assert-lot`)
	if p.Parse() == nil {
		t.Errorf("assert-lot function succeeded but should have failed")
	}
}

func TestAssertLotFunction_NonexistentCommodity(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		JPY Yen commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 1 JPY xfer foolot create-lot
			Equity -1 JPY xfer
			xact
		Assets:Account foolot 0 USD assert-lot`)
	if p.Parse() == nil {
		t.Errorf("assert-lot function succeeded but should have failed")
	}
}

func TestAssertLotFunction_WrongCommodity(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		JPY Yen commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 1 JPY xfer foolot create-lot
			Equity -1 JPY xfer
			xact
		Assets:Account foolot 1 USD assert-lot`)
	if p.Parse() == nil {
		t.Errorf("assert-lot function succeeded but should have failed")
	}
}

func TestAssertLotFunction_ClosedAccount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Assets:Account close
		Assets:Account foolot 0 USD assert-lot`)
	if p.Parse() == nil {
		t.Errorf("assert-lot function succeeded but should have failed")
	}
}

func TestAssertLotsSumFunction(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		JPY Yen commodity
		Assets:Account open
		Equity open
		(Entity Description
			Assets:Account 10 USD xfer
			Assets:Account -20 USD xfer foolot create-lot
			Assets:Account 30 USD xfer barlot create-lot
			Assets:Account 10 JPY 1 USD 10 USD xfer-exch barlot create-lot
			Equity -15 USD xfer
			Equity -5 USD xfer barlot create-lot
			Equity -10 JPY 1 USD -10 USD xfer-exch barlot create-lot
			xact)
		Assets:Account 20 USD assert-lots-sum
		Assets:Account 10 JPY assert-lots-sum
		Equity -20 USD assert-lots-sum
		Equity -10 JPY assert-lots-sum)`)
	if e := p.Parse(); e != nil {
		t.Errorf("assert-lots-sum function failed: %v", e)
	}
}

func TestAssertLotsSumFunction_WrongAmount(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 10 USD xfer foolot create-lot
			Equity -10 USD xfer
			xact
		Assets:Account 10.1 USD assert-lots-sum)`)
	if p.Parse() == nil {
		t.Errorf("assert-lots-sum function succeeded but should have failed")
	}
}

func TestAssertLotsSumFunction_NonzeroAmountOfAbsentCommodity(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Assets:Account 1 USD assert-lots-sum)`)
	if p.Parse() == nil {
		t.Errorf("assert-lots-sum function succeeded but should have failed")
	}
}

func TestAssertLotsSumFunction_ZeroAmountOfAbsentCommodity(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		JPY Yen commodity
		Assets:Account open
		Equity open
		(Entity Description
			Assets:Account 1 JPY xfer foolot create-lot
			Equity -1 JPY xfer
			xact)
		Assets:Account 0 USD assert-lots-sum)`)
	if e := p.Parse(); e != nil {
		t.Errorf("assert-lots-sum function failed: %v", e)
	}
}

func TestAssertLotsSumFunction_TooFewOperands(t *testing.T) {
	for _, program := range []string{"assert-lots-sum", "Assets:Account assert-lots-sum", "Assets:Account 1 assert-lots-sum"} {
		p := createParser(program)
		if p.Parse() == nil {
			t.Errorf("assert-lots-sum function succeeded but should have failed")
		}
	}
}

func TestAssertLotsSumFunction_NonStringAccountName(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		123 atoi 0 USD assert-lots-sum`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("assert-lots-sum function succeeded but should have failed")
	}
}

func TestAssertLotsSumFunction_IllegalAmount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Assets:Account 0a USD assert-lots-sum`)
	if p.Parse() == nil {
		t.Errorf("assert-lots-sum function succeeded but should have failed")
	}
}

func TestAssertLotsSumFunction_NonStringCommodityName(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account 0 123 atoi assert-lots-sum`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("assert-lots-sum function succeeded but should have failed")
	}
}

func TestAssertLotsSumFunction_NonexistentAccount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account 0 USD assert-lots-sum`)
	if p.Parse() == nil {
		t.Errorf("assert-lots-sum function succeeded but should have failed")
	}
}

func TestAssertLotsSumFunction_NonexistentCommodity(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		JPY Yen commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 1 JPY xfer
			Equity -1 JPY xfer
			xact
		Assets:Account 0 USD assert-lots-sum`)
	if p.Parse() == nil {
		t.Errorf("assert-lots-sum function succeeded but should have failed")
	}
}

func TestAssertLotsSumFunction_WrongCommodity(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		JPY Yen commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 1 JPY xfer
			Equity -1 JPY xfer
			xact
		Assets:Account 1 USD assert-lots-sum`)
	if p.Parse() == nil {
		t.Errorf("assert-lots-sum function succeeded but should have failed")
	}
}

func TestAssertLotsSumFunction_ClosedAccount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Assets:Account close
		Assets:Account 0 USD assert-lots-sum`)
	if p.Parse() == nil {
		t.Errorf("assert-lots-sum function succeeded but should have failed")
	}
}

func TestCloseFunction(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account close`)
	if e := p.Parse(); e != nil {
		t.Errorf("close function failed: %v", e)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf("open did not create an account in the Context")
	} else if a.Name != "Assets:Account" {
		t.Errorf("open created an account with the wrong name: %v", a.Name)
	} else if !a.IsClosed(p.Context().Date) {
		t.Errorf("close did not close the account, closing date is %v", a.ClosingDate)
	}
}

func TestCloseFunction_ZeroOperands(t *testing.T) {
	p := createParser(`close`)
	if p.Parse() == nil {
		t.Errorf("close function should have failed but succeeded")
	}
}

func TestCloseFunction_NonStringAccountNameOperand(t *testing.T) {
	p := createParser(`123 atoi close`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`close function succeeded with non-string account name`)
	}
}

func TestCloseFunction_NonexistentAccount(t *testing.T) {
	p := createParser(`date 2000 1 1 Assets:Account close`)
	if p.Parse() == nil {
		t.Errorf("close function should have failed but succeeded")
	}
}

func TestCloseFunction_AccountAlreadyClosed(t *testing.T) {
	p := createParser(`
		date 2000 1 1
		Assets:Account open
		Assets:Account close
		Assets:Account close`)
	if p.Parse() == nil {
		t.Errorf("close function should have failed but succeeded")
	}
}

func TestCloseFunction_AccountHasNonzeroLots(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Equity open
		USD Dollar commodity
		Entity Description
			Assets:Account 20 USD xfer foo lot
			Equity -20 USD xfer
			xact
		Assets:Account close`)
	if p.Parse() == nil {
		t.Errorf("close function should have failed but succeeded")
	}
}

func TestCloseFunction_DefaultLotIsNonzero(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Equity open
		USD Dollar commodity
		Entity Description
			Assets:Account 20 USD xfer
			Equity -20 USD xfer
			xact
		Assets:Account close`)
	if err := p.Parse(); err != nil {
		t.Errorf("close function failed: %v", err)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf("open did not create an account in the Context")
	} else if a.Name != "Assets:Account" {
		t.Errorf("open created an account with the wrong name: %v", a.Name)
	} else if !a.IsClosed(p.Context().Date) {
		t.Errorf("close did not close the account, closing date is %v", a.ClosingDate)
	} else if len(a.Lots) != 1 {
		t.Errorf("Assets:Account has %v lots instead of 1", len(a.Lots))
	}
}

func TestCloseFunction_AccountHasOnlyLotsWithZeroBalances(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Equity open
		USD Dollar commodity
		JPY Yen commodity
		Entity Description
			Assets:Account 20 USD xfer
			Equity -20 USD xfer
			xact
		Entity Description
			Assets:Account 30 JPY xfer foolot create-lot
			Equity -30 JPY xfer
			xact
		Entity Description
			Assets:Account -20 USD xfer
			Equity 20 USD xfer
			xact
		Entity Description
			Assets:Account -30 JPY xfer foolot lot
			Equity 30 JPY xfer
			xact
		Assets:Account close`)
	if err := p.Parse(); err != nil {
		t.Errorf("close function failed: %v", err)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf("open did not create an account in the Context")
	} else if a.Name != "Assets:Account" {
		t.Errorf("open created an account with the wrong name: %v", a.Name)
	} else if !a.IsClosed(p.Context().Date) {
		t.Errorf("close did not close the account, closing date is %v", a.ClosingDate)
	}
}

func TestCloseLotFunction(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Equity open
		USD Dollar commodity
		Entity Description
			Assets:Account 1 USD xfer
			Assets:Account 2 USD xfer foolot create-lot
			Equity -3 USD xfer
			xact
		Entity Description
			Assets:Account -2 USD xfer foolot lot
			Equity 2 USD xfer
			xact
		Assets:Account foolot close-lot`)
	if err := p.Parse(); err != nil {
		t.Errorf(`close-lot function failed: %v`, err)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf("open did not create an account in the Context")
	} else if a.IsClosed(p.Context().Date) {
		t.Errorf("close-lot closed the account instead of the lot")
	} else if _, ok := a.Lots["foolot"]; ok {
		t.Errorf("close-lot did not delete the lot")
	} else if ctol, ok := a.Lots[""]; !ok {
		t.Errorf("close-lot deleted the wrong lot (the default lot)")
	} else if l, ok := ctol["USD"]; !ok {
		t.Errorf("default lot does not have USD")
	} else if !l.Balance.Amount.Equal(decimal.NewFromInt(1)) {
		t.Errorf("default lot's balance is not 1 USD: %v", &l.Balance)
	}
}

func TestCloseLotFunction_ZeroOperands(t *testing.T) {
	p := createParser(`close-lot`)
	if p.Parse() == nil {
		t.Errorf(`close-lot function succeeded with zero operands`)
	}
}

func TestCloseLotFunction_NoLotNameOperands(t *testing.T) {
	p := createParser(`Assets:Account open Assets:Account close-lot`)
	if p.Parse() == nil {
		t.Errorf(`close-lot function succeeded with no lot name operand`)
	}
}

func TestCloseLotFunction_NonStringAccountNameOperand(t *testing.T) {
	p := createParser(`123 atoi foolot close-lot`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`close-lot function succeeded with non-string account name`)
	}
}

func TestCloseLotFunction_NonStringLotNameOperand(t *testing.T) {
	p := createParser(`Assets:Account open Assets:Account 123 atoi close-lot`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`close-lot function succeeded with non-string lot name`)
	}
}

func TestCloseLotFunction_NonexistentAccount(t *testing.T) {
	p := createParser(`Assets:Account foolot close-lot`)
	if p.Parse() == nil {
		t.Errorf(`close-lot function succeeded with nonexistent account and lot`)
	}
}

func TestCloseLotFunction_AccountIsClosed(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account close
		Assets:Account "" close-lot`)
	if p.Parse() == nil {
		t.Errorf(`close-lot function succeeded with closed account`)
	}
}

func TestCloseLotFunction_NonexistentLot(t *testing.T) {
	p := createParser(`Assets:Account open Assets:Account foolot close-lot`)
	if p.Parse() == nil {
		t.Errorf(`close-lot function succeeded with nonexistent lot`)
	}
}

func TestCloseLotFunction_LotHasANonzeroBalance(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Equity open
		USD Dollar commodity
		JPY Yen commodity
		Entity Description
			Assets:Account 1 USD xfer
			Assets:Account 2 USD xfer foolot create-lot
			Assets:Account 3 JPY 1 USD 3 USD xfer-exch foolot create-lot
			Equity -6 USD xfer
			xact
		Entity Description
			Assets:Account -2 USD xfer foolot lot
			Equity 2 USD xfer
			xact
		Assets:Account foolot close-lot`)
	if e := p.Parse(); e == nil {
		t.Errorf(`close-lot function succeeded when a lot had a nonzero balance`)
	}
}

func TestCommentFunction_OneStringOperand(t *testing.T) {
	p := createParser(`"This is a comment." comment`)
	if e := p.Parse(); e != nil {
		t.Errorf("comment function failed: %v", e)
	}
}

func TestCommentFunction_ZeroOperands(t *testing.T) {
	p := createParser(`comment`)
	if p.Parse() == nil {
		t.Errorf("comment function succeeded but should have failed")
	}
}

func TestCommentFunction_TwoStringOperands(t *testing.T) {
	p := createParser(`a b comment`)
	if p.Parse() == nil {
		t.Errorf("program succeeded but should have failed")
	}
}

func TestCommentFunction_NonStringOperand(t *testing.T) {
	p := createParser(`12345 atoi comment`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("comment function succeeded but should have failed")
	}
}

func TestCommodityFunction_TwoDifferentCommodities(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD "United States Dollar" commodity
		2011 3 11 date
		JPY "Japanese Yen" commodity`)
	if e := p.Parse(); e != nil {
		t.Errorf("commodity function failed: %v", e)
	}
	var c *core.Commodity
	var ok bool
	if c, ok = p.Context().Commodities["USD"]; !ok {
		t.Errorf("commodity did not create USD")
	} else if c.Name != "USD" {
		t.Errorf("commodity did not set commodity name to USD")
	} else if c.Description != "United States Dollar" {
		t.Errorf("commodity did not set description to United States Dollar")
	} else if !reflect.DeepEqual(c.CreationDate, core.Date{2000, 1, 1}) {
		t.Errorf("commodity did not use current date")
	}
	if c, ok = p.Context().Commodities["JPY"]; !ok {
		t.Errorf("commodity did not create JPY")
	} else if c.Name != "JPY" {
		t.Errorf("commodity did not set commodity name to JPY")
	} else if c.Description != "Japanese Yen" {
		t.Errorf("commodity did not set description to Japanese Yen")
	} else if !reflect.DeepEqual(c.CreationDate, core.Date{2011, 3, 11}) {
		t.Errorf("commodity did not use current date")
	}
}

func TestCommodityFunction_TooFewOperands(t *testing.T) {
	for _, program := range []string{"commodity", "USD commodity"} {
		p := createParser(program)
		if p.Parse() == nil {
			t.Errorf(`"%v" succeeded but should have failed`, program)
		}
	}
}

func TestCommodityFunction_NonStringCommodityName(t *testing.T) {
	p := createParser(`12345 atoi "United States Dollar" commodity`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("commodity should have failed but succeeded")
	}
}

func TestCommodityFunction_NonStringDescription(t *testing.T) {
	p := createParser(`USD 12345 atoi commodity`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("commodity should have failed but succeeded")
	}
}

func TestCommodityFunction_ExistingCommodity(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD "Dollar" commodity
		USD "Duplicate" commodity`)
	if p.Parse() == nil {
		t.Errorf("commodity should have failed but succeeded")
	}
}

func TestCreateLotFunction_LotExistsWithCommodity(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 1 USD xfer foolot create-lot
			Equity -1 USD xfer
			xact
		Assets:Account 1 USD xfer foolot create-lot`)
	if p.Parse() == nil {
		t.Errorf("create-lot should have failed but succeeded")
	}
}

func TestCreateLotFunction_LotExistsWithoutCommodity(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		JPY Yen commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 1 USD xfer foolot create-lot
			Equity -1 USD xfer
			xact
		Entity Description
			Assets:Account 2 JPY xfer foolot create-lot
			Equity -2 JPY xfer
			xact)`)
	if e := p.Parse(); e != nil {
		t.Errorf("create-lot function failed: %v", e)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf("open did not create an account")
	} else if ctol, ok := a.Lots["foolot"]; !ok {
		t.Errorf("create-lot did not create a lot")
	} else if l, ok := ctol["USD"]; !ok {
		t.Errorf("create-lot did not create USD lot")
	} else if l.Name != "foolot" {
		t.Errorf("create-lot did not set correct lot name, got %v", l.Name)
	} else if !reflect.DeepEqual(l.CreationDate, core.Date{2000, 1, 1}) {
		t.Errorf("create-lot did not set correct creation date, got %v", l.CreationDate)
	} else if l.Balance.Commodity == nil || l.Balance.Commodity.Name != "USD" {
		t.Errorf("create-lot did not set correct commodity, got %v", l.Balance)
	} else if !decimal.NewFromInt(1).Equal(l.Balance.Amount) {
		t.Errorf("create-lot did not set correct amount, got %v", l.Balance.Amount)
	} else if l, ok := ctol["JPY"]; !ok {
		t.Errorf("create-lot did not create JPY lot")
	} else if l.Name != "foolot" {
		t.Errorf("create-lot did not set correct lot name, got %v", l.Name)
	} else if !reflect.DeepEqual(l.CreationDate, core.Date{2000, 1, 1}) {
		t.Errorf("create-lot did not set correct creation date, got %v", l.CreationDate)
	} else if l.Balance.Commodity == nil || l.Balance.Commodity.Name != "JPY" {
		t.Errorf("create-lot did not set correct commodity, got %v", l.Balance)
	} else if !decimal.NewFromInt(2).Equal(l.Balance.Amount) {
		t.Errorf("create-lot did not set correct amount, got %v", l.Balance.Amount)
	}
}

func TestCreateLotFunction_WithXfer(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 1 USD xfer foolot create-lot
			Equity -1 USD xfer
			xact)`)
	if e := p.Parse(); e != nil {
		t.Errorf("create-lot function failed: %v", e)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf("open did not create an account")
	} else if ctol, ok := a.Lots["foolot"]; !ok {
		t.Errorf("create-lot did not create a lot")
	} else if l, ok := ctol["USD"]; !ok {
		t.Errorf("create-lot did not create USD lot")
	} else if l.Name != "foolot" {
		t.Errorf("create-lot did not set correct lot name, got %v", l.Name)
	} else if !reflect.DeepEqual(l.CreationDate, core.Date{2000, 1, 1}) {
		t.Errorf("create-lot did not set correct creation date, got %v", l.CreationDate)
	} else if l.Balance.Commodity == nil || l.Balance.Commodity.Name != "USD" {
		t.Errorf("create-lot did not set correct commodity, got %v", l.Balance)
	} else if !decimal.NewFromInt(1).Equal(l.Balance.Amount) {
		t.Errorf("create-lot did not set correct amount, got %v", l.Balance.Amount)
	}
}

func TestCreateLotFunction_WithXferExch(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		JPY Yen commodity
		Assets:Account open
		Equity open
		Entity Description
			Assets:Account 2 USD 100 JPY 200 JPY xfer-exch foolot create-lot
			Equity -200 JPY xfer
			xact)`)
	if e := p.Parse(); e != nil {
		t.Errorf("create-lot function failed: %v", e)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf("open did not create an account")
	} else if ctol, ok := a.Lots["foolot"]; !ok {
		t.Errorf("create-lot did not create a lot")
	} else if l, ok := ctol["USD"]; !ok {
		t.Errorf("create-lot did not create USD lot")
	} else if l.Name != "foolot" {
		t.Errorf("create-lot did not set correct lot name, got %v", l.Name)
	} else if !reflect.DeepEqual(l.CreationDate, core.Date{2000, 1, 1}) {
		t.Errorf("create-lot did not set correct creation date, got %v", l.CreationDate)
	} else if l.Balance.Commodity == nil || l.Balance.Commodity.Name != "USD" {
		t.Errorf("create-lot did not set correct commodity, got %v", l.Balance)
	} else if !decimal.NewFromInt(2).Equal(l.Balance.Amount) {
		t.Errorf("create-lot did not set correct amount, got %v", l.Balance.Amount)
	} else if l.ExchangeRate == nil {
		t.Errorf("create-lot did not set exchange rate")
	} else if l.ExchangeRate.UnitPrice.Commodity == nil || l.ExchangeRate.UnitPrice.Commodity.Name != "JPY" {
		t.Errorf("create-lot did not set correct unit price commodity, got %v", l.ExchangeRate.UnitPrice.Commodity)
	} else if !decimal.NewFromInt(100).Equal(l.ExchangeRate.UnitPrice.Amount) {
		t.Errorf("create-lot did not set correct unit price amount, got %v", l.ExchangeRate.UnitPrice.Amount)
	} else if l.ExchangeRate.TotalPrice.Commodity == nil || l.ExchangeRate.TotalPrice.Commodity.Name != "JPY" {
		t.Errorf("create-lot did not set correct total price commodity, got %v", l.ExchangeRate.TotalPrice.Commodity)
	} else if !decimal.NewFromInt(200).Equal(l.ExchangeRate.TotalPrice.Amount) {
		t.Errorf("create-lot did not set correct total price amount, got %v", l.ExchangeRate.TotalPrice.Amount)
	}
}

func TestDateFunction_ValidDateSequence(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		2000 1 2 date
		2001 9 11 date`)
	if e := p.Parse(); e != nil {
		t.Errorf("date function failed: %v", e)
	}
}

func TestDateFunction_NotEnoughOperands(t *testing.T) {
	for _, program := range []string{"date", "2000 date", "2000 1 date"} {
		p := createParser(program)
		if p.Parse() == nil {
			t.Errorf(`"%v" succeeded but should have failed`, program)
		}
	}
}

func TestDateFunction_NonStringYear(t *testing.T) {
	p := createParser(`2000 atoi 1 1 date`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("date succeeded but should have failed")
	}
}

func TestDateFunction_NonStringMonth(t *testing.T) {
	p := createParser(`2000 1 atoi 1 date`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("date succeeded but should have failed")
	}
}

func TestDateFunction_NonStringDay(t *testing.T) {
	p := createParser(`2000 1 1 atoi date`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("date succeeded but should have failed")
	}
}

func TestDateFunction_InvalidYear(t *testing.T) {
	p := createParser(`2000a 1 1 date`)
	if p.Parse() == nil {
		t.Errorf("date succeeded but should have failed")
	}
}

func TestDateFunction_InvalidMonth(t *testing.T) {
	p := createParser(`2000 1b 1 date`)
	if p.Parse() == nil {
		t.Errorf("date succeeded but should have failed")
	}
}

func TestDateFunction_InvalidDay(t *testing.T) {
	p := createParser(`2000 1 1c date`)
	if p.Parse() == nil {
		t.Errorf("date succeeded but should have failed")
	}
}

func TestDateFunction_DateGoesBackwardsInTime(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		1999 12 31 date`)
	if p.Parse() == nil {
		t.Errorf("date succeeded but should have failed")
	}
}

func TestLotFunctions(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Equity open)
		Entity Description
			Assets:Account 20 USD xfer foolot create-lot
			Equity -20 USD xfer
			xact
		Entity Description
			Assets:Account -5 USD xfer foolot lot
			Equity 5 USD xfer
			xact`)
	if err := p.Parse(); err != nil {
		t.Errorf(`one of the lot functions failed: %v`, err)
	} else if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf(`Assets:Account does not exist`)
	} else if len(a.Lots) != 2 {
		t.Errorf(`Assets:Account has %v lots instead of 2`, len(a.Lots))
	} else if ctol, ok := a.Lots["foolot"]; !ok {
		t.Errorf(`Assets:Account does not have a foolot lot`)
	} else if l, ok := ctol["USD"]; !ok {
		t.Errorf(`foolot does not have USD`)
	} else if !l.Balance.Amount.Equal(decimal.NewFromInt(15)) {
		t.Errorf(`foolot has %v USD instead of 15`, l.Balance.Amount)
	}
}

func TestLotFunction_TooFewArgs(t *testing.T) {
	for _, prog := range []string{`lot`, `foolot lot`} {
		if createParser(prog).Parse() == nil {
			t.Errorf(`program succeeded but should have failed: %v`, prog)
		}
	}
}

func TestLotFunction_NonTransferOperand(t *testing.T) {
	if createParser(`Assets:Account foolot lot`).Parse() == nil {
		t.Errorf(`program succeeded but should have failed`)
	}
}

func TestLotFunction_NonStringLotNameOperand(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Equity open)
		Entity Description
			Assets:Account 5 USD xfer 123 atoi lot
			Equity -5 USD xfer
			xact`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`program succeeded but should have failed`)
	}
}

func TestLotFunction_LotDoesNotExist(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Equity open)
		Entity Description
			Assets:Account 5 USD xfer foolot lot
			Equity -5 USD xfer
			xact`)
	if p.Parse() == nil {
		t.Errorf(`program succeeded but should have failed`)
	}
}

func TestLotFunction_LotExistsWithAnotherCommodity(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		JPY Yen commodity
		Assets:Account open
		Equity open)
		Entity Description
			Assets:Account 20 JPY xfer foolot create-lot
			Equity -20 JPY xfer
			xact
		Entity Description
			Assets:Account 5 USD xfer foolot lot
			Equity -5 USD xfer
			xact`)
	if err := p.Parse(); err != nil {
		t.Errorf(`one of the lot functions failed: %v`, err)
	} else if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf(`Assets:Account does not exist`)
	} else if len(a.Lots) != 2 {
		t.Errorf(`Assets:Account has %v lots instead of 2`, len(a.Lots))
	} else if ctol, ok := a.Lots["foolot"]; !ok {
		t.Errorf(`Assets:Account does not have a foolot lot`)
	} else if len(ctol) != 2 {
		t.Errorf(`foolot has %v commodities instead of 2`, len(ctol))
	} else if l, ok := ctol["USD"]; !ok {
		t.Errorf(`foolot does not have USD`)
	} else if !l.Balance.Amount.Equal(decimal.NewFromInt(5)) {
		t.Errorf(`foolot has %v USD instead of 5`, l.Balance.Amount)
	} else if l, ok := ctol["JPY"]; !ok {
		t.Errorf(`foolot does not have JPY`)
	} else if !l.Balance.Amount.Equal(decimal.NewFromInt(20)) {
		t.Errorf(`foolot has %v USD instead of 20`, l.Balance.Amount)
	}
}

func TestOpenFunction(t *testing.T) {
	p := createParser(`2000 1 1 date Assets:Account open`)
	if err := p.Parse(); err != nil {
		t.Errorf("open failed: %v", err)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf("open did not create an account in the Context")
	} else if a.Name != "Assets:Account" {
		t.Errorf("open created an account with the wrong name: %v", a.Name)
	} else if a.CreationDate != p.Context().Date {
		t.Errorf("open created an account with the wrong creation date: %v", a.CreationDate)
	} else if !reflect.DeepEqual(a.CreationDate, core.Date{2000, 1, 1}) {
		t.Errorf("open did not use current date")
	} else if a.IsClosed(p.Context().Date) {
		t.Errorf("open created an account closed on %v", a.ClosingDate)
	} else if len(a.Commodities) != 0 {
		t.Errorf("open created an account with commodity limitations: %v", a.Commodities)
	} else if len(a.Lots) != 1 {
		t.Errorf("open created an account with %v lots instead of the default one: %v", len(a.Lots), a.Lots)
	} else if dl, ok := a.Lots[""]; !ok {
		t.Errorf("open created an account without a default lot")
	} else if len(dl) != 0 {
		t.Errorf("open created an account with a nonempty default lot: %v", dl)
	} else if len(a.GetTags()) != 0 {
		t.Errorf("open created an account with tags: %v", a.GetTags())
	}
}

func TestOpenFunction_WithCommodities(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		JPY Yen commodity
		Assets:Account USD JPY open`)
	if err := p.Parse(); err != nil {
		t.Errorf("open failed: %v", err)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf("open did not create an account in the Context")
	} else if a.Name != "Assets:Account" {
		t.Errorf("open created an account with the wrong name: %v", a.Name)
	} else if a.CreationDate != p.Context().Date {
		t.Errorf("open created an account with the wrong creation date: %v", a.CreationDate)
	} else if !reflect.DeepEqual(a.CreationDate, core.Date{2000, 1, 1}) {
		t.Errorf("open did not use current date")
	} else if a.IsClosed(p.Context().Date) {
		t.Errorf("open created an account closed on %v", a.ClosingDate)
	} else if len(a.Commodities) != 2 {
		t.Errorf("open created an account with other than two commodity limitations: %v", a.Commodities)
	} else if c, ok := a.Commodities["USD"]; !ok {
		t.Errorf("open created an account without commodity limitation USD")
	} else if c.Name != "USD" {
		t.Errorf("open created an account with commodity limitation USD, but points to commodity %v", c.Name)
	} else if c, ok := a.Commodities["JPY"]; !ok {
		t.Errorf("open created an account without commodity limitation JPY")
	} else if c.Name != "JPY" {
		t.Errorf("open created an account with commodity limitation USD, but points to commodity %v", c.Name)
	} else if len(a.Lots) != 1 {
		t.Errorf("open created an account with %v lots instead of the default one: %v", len(a.Lots), a.Lots)
	} else if dl, ok := a.Lots[""]; !ok {
		t.Errorf("open created an account without a default lot")
	} else if len(dl) != 0 {
		t.Errorf("open created an account with a nonempty default lot: %v", dl)
	} else if len(a.GetTags()) != 0 {
		t.Errorf("open created an account with tags: %v", a.GetTags())
	}
}

func TestOpenFunction_ValidPrefixes(t *testing.T) {
	p := createParser(`
		Assets:Foo open
		Liabilities:Foo open
		Income:Foo open
		Expenses:Foo open
		Equity:Foo open
		Equity open`)
	if err := p.Parse(); err != nil {
		t.Errorf(`open failed: %v`, err)
	} else if len(p.Context().Accounts) != 6 {
		t.Errorf(`did not open six accounts: %v`, p.Context().Accounts)
	}
}

func TestOpenFunction_InvalidAccountName(t *testing.T) {
	p := createParser(`foobar open`)
	if p.Parse() == nil {
		t.Errorf(`open succeeded with an invalid account name`)
	}
}

func TestOpenFunction_ZeroOperands(t *testing.T) {
	p := createParser(`open`)
	if p.Parse() == nil {
		t.Errorf("open succeeded but should have failed")
	}
}

func TestOpenFunction_NonStringAccountName(t *testing.T) {
	p := createParser(`123 atoi open`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`open succeeded with non-string account name`)
	}
}

func TestOpenFunction_NonexistentCommodity(t *testing.T) {
	p := createParser(`
		USD Dollar commodity
		Assets:Account USD NONEXISTENT open`)
	if p.Parse() == nil {
		t.Errorf("open succeeded but should have failed")
	}
}

func TestOpenFunction_ExistingOpenAccount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account open`)
	if p.Parse() == nil {
		t.Errorf("open succeeded but should have failed")
	}
	p = createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account USD open
		Assets:Account USD open`)
	if p.Parse() == nil {
		t.Errorf("open succeeded but should have failed")
	}
	p = createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		Assets:Account USD open`)
	if p.Parse() == nil {
		t.Errorf("open succeeded but should have failed")
	}
	p = createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account USD open
		Assets:Account open`)
	if p.Parse() == nil {
		t.Errorf("open succeeded but should have failed")
	}
}

func TestOpenFunction_ClosedAccount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		USD Dollar commodity
		Assets:Account open
		2000 1 2 date
		Assets:Account close
		2000 1 3 date
		Assets:Account USD open`)
	if err := p.Parse(); err != nil {
		t.Errorf("open failed: %v", err)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf("open did not create an account in the Context")
	} else if a.Name != "Assets:Account" {
		t.Errorf("open created an account with the wrong name: %v", a.Name)
	} else if a.CreationDate != p.Context().Date {
		t.Errorf("open created an account with the wrong creation date: %v", a.CreationDate)
	} else if !reflect.DeepEqual(a.CreationDate, core.Date{2000, 1, 3}) {
		t.Errorf("open did not use current date")
	} else if a.IsClosed(p.Context().Date) {
		t.Errorf("open created an account closed on %v", a.ClosingDate)
	} else if len(a.Commodities) != 1 {
		t.Errorf("open created an account with other than two commodity limitations: %v", a.Commodities)
	} else if c, ok := a.Commodities["USD"]; !ok {
		t.Errorf("open created an account without commodity limitation USD")
	} else if c.Name != "USD" {
		t.Errorf("open created an account with commodity limitation USD, but points to commodity %v", c.Name)
	} else if len(a.Lots) != 1 {
		t.Errorf("open created an account with %v lots instead of the default one: %v", len(a.Lots), a.Lots)
	} else if dl, ok := a.Lots[""]; !ok {
		t.Errorf("open created an account without a default lot")
	} else if len(dl) != 0 {
		t.Errorf("open created an account with a nonempty default lot: %v", dl)
	} else if len(a.GetTags()) != 0 {
		t.Errorf("open created an account with tags: %v", a.GetTags())
	}
}

func TestSetCommentFunction(t *testing.T) {
	checkComment := func(fn string, op parser.Operands, ctx *core.Context) error {
		if op.Length() != 1 {
			t.Errorf("set-comment did not leave exactly one operand on the stack, left %v", op.Length())
			return fmt.Errorf("test failed")
		}
		values := op.Pop(1)
		if xfer, ok := values[0].(*Transfer); !ok {
			t.Errorf("set-comment did not push a *Transfer onto the stack, pushed %v", values[0])
			return fmt.Errorf("test failed")
		} else if xfer.Comment != "test comment" {
			t.Errorf("set-comment did not set the Transfer's comment correctly, set: %v", xfer.Comment)
			return fmt.Errorf("test failed")
		}
		return nil
	}
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		Assets:Account open)
		Assets:Account 5 USD xfer
		"test comment" set-comment
		test-check-comment`)
	p.Functions["test-check-comment"] = checkComment
	if e := p.Parse(); e != nil {
		t.Errorf("set-comment failed: %v", e)
	}
}

func TestSetCommentFunction_TooFewOperands(t *testing.T) {
	for _, prog := range []string{`set-comment`, `Assets:Account set-comment`} {
		p := createParser(prog)
		if p.Parse() == nil {
			t.Errorf("set-comment succeeded but should have failed for program: %v", prog)
		}
	}
}

func TestSetCommentFunction_NonTransferOperand(t *testing.T) {
	p := createParser(`"foo transfer" "overwritten comment" set-comment`)
	if p.Parse() == nil {
		t.Errorf("set-comment succeeded but should have failed")
	}
}

func TestSetCommentFunction_NonStringComment(t *testing.T) {
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		Assets:Account open)
		Assets:Account 5 USD xfer
		123 atoi set-comment`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf("set-comment succeeded but should have failed")
	}
}

func TestSetCommentFunction_Repeated(t *testing.T) {
	checkComment := func(fn string, op parser.Operands, ctx *core.Context) error {
		if op.Length() != 1 {
			t.Errorf("set-comment did not leave exactly one operand on the stack, left %v", op.Length())
			return fmt.Errorf("test failed")
		}
		values := op.Pop(1)
		if xfer, ok := values[0].(*Transfer); !ok {
			t.Errorf("set-comment did not push a *Transfer onto the stack, pushed %v", values[0])
			return fmt.Errorf("test failed")
		} else if xfer.Comment != "test comment" {
			t.Errorf("set-comment did not set the Transfer's comment correctly, set: %v", xfer.Comment)
			return fmt.Errorf("test failed")
		}
		return nil
	}
	p := createParser(`
		(2000 1 1 date
		USD Dollar commodity
		Assets:Account open)
		Assets:Account 5 USD xfer
		"overwritten comment" set-comment
		"test comment" set-comment
		test-check-comment`)
	p.Functions["test-check-comment"] = checkComment
	if e := p.Parse(); e != nil {
		t.Errorf("set-comment failed: %v", e)
	}
}

func TestTagFunction(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account foo bar tag`)
	if err := p.Parse(); err != nil {
		t.Errorf(`tag failed: %v`, err)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf(`open did not create an account in the Context`)
	} else if len(a.GetTags()) != 2 {
		t.Errorf(`the account does not have two tags, it has %v`, len(a.GetTags()))
	} else if !a.HasTag("foo") {
		t.Errorf(`the account is not tagged with "foo"`)
	} else if !a.HasTag("bar") {
		t.Errorf(`the account is not tagged with "bar"`)
	}
	for _, tag := range []string{"foo", "bar"} {
		if tagged, ok := p.Context().Tags[tag]; !ok {
			t.Errorf(`the Context does not have a "%v" tag`, tag)
		} else if len(tagged) != 1 {
			t.Errorf(`the "%v" tag does not have exactly one object`, tag)
		} else if a, ok := tagged[0].(*core.Account); !ok {
			t.Errorf(`the object tagged with "%v" is not an Account`, tag)
		} else if a != p.Context().Accounts["Assets:Account"] {
			t.Errorf(`the object tagged with "%v" is not the Assets:Account account: %v`, tag, a.Name)
		}
	}
}

func TestTagFunction_ZeroOperands(t *testing.T) {
	p := createParser(`tag`)
	if p.Parse() == nil {
		t.Errorf(`tag succeeded with zero operands`)
	}
}

func TestTagFunction_NoTagOperands(t *testing.T) {
	p := createParser(`Assets:Account open Assets:Account tag`)
	if p.Parse() == nil {
		t.Errorf(`tag succeeded with no tag operands`)
	}
}

func TestTagFunction_NonStringAccountNameOperand(t *testing.T) {
	p := createParser(`123 atoi foo tag`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`tag succeeded with a non-string account name operand`)
	}
}

func TestTagFunction_NonStringTagOperand(t *testing.T) {
	p := createParser(`Assets:Account open Assets:Account 123 atoi tag`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`tag succeeded with a non-string tag operand`)
	}
}

func TestTagFunction_NonexitentAccount(t *testing.T) {
	p := createParser(`Assets:Account foo tag`)
	if p.Parse() == nil {
		t.Errorf(`tag succeeded with a nonexistent account`)
	}
}

func TestTagFunction_ClosedAccount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account close
		Assets:Account foo tag`)
	if p.Parse() == nil {
		t.Errorf(`tag succeeded with a closed account`)
	}
}

func TestTagFunction_DuplicateTags(t *testing.T) {
	p := createParser(`Assets:Account open Assets:Account foo bar foo tag`)
	if err := p.Parse(); err != nil {
		t.Errorf(`tag failed: %v`, err)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf(`open did not create an account in the Context`)
	} else if len(a.GetTags()) != 2 {
		t.Errorf(`the account does not have two tags, it has %v`, len(a.GetTags()))
	} else if !a.HasTag("foo") {
		t.Errorf(`the account is not tagged with "foo"`)
	} else if !a.HasTag("bar") {
		t.Errorf(`the account is not tagged with "bar"`)
	}
	for _, tag := range []string{"foo", "bar"} {
		if tagged, ok := p.Context().Tags[tag]; !ok {
			t.Errorf(`the Context does not have a "%v" tag`, tag)
		} else if len(tagged) != 1 {
			t.Errorf(`the "%v" tag does not have exactly one object`, tag)
		} else if a, ok := tagged[0].(*core.Account); !ok {
			t.Errorf(`the object tagged with "%v" is not an Account`, tag)
		} else if a != p.Context().Accounts["Assets:Account"] {
			t.Errorf(`the object tagged with "%v" is not the Assets:Account account: %v`, tag, a.Name)
		}
	}
}

func TestTagFunction_TwoAccounts(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Foo open
		Assets:Bar open
		Assets:Foo foo tag
		Assets:Bar foo tag`)
	if err := p.Parse(); err != nil {
		t.Errorf(`tag failed: %v`, err)
	}
	if tagged, ok := p.Context().Tags["foo"]; !ok {
		t.Errorf(`the Context does not have a "foo" tag`)
	} else if len(tagged) != 2 {
		t.Errorf(`the "foo" tag does not have two objects, it has %v`, len(tagged))
	} else {
		for _, an := range []string{"Assets:Foo", "Assets:Bar"} {
			if a, ok := p.Context().Accounts[an]; !ok {
				t.Errorf(`open did not create an account named %v in the Context`, an)
			} else if len(a.GetTags()) != 1 {
				t.Errorf(`the %v account does not have one tag, it has %v`, an, len(a.GetTags()))
			} else if !a.HasTag("foo") {
				t.Errorf(`the %v account is not tagged with "foo"`, an)
			} else {
				found := false
				for _, to := range tagged {
					if to == a {
						found = true
						break
					}
				}
				if !found {
					t.Errorf(`the %v account is not in Context.Tags["foo"]`, an)
				}
			}
		}
	}
}

func TestTagCommodityFunction(t *testing.T) {
	p := createParser(`USD Dollar commodity USD foo bar tag-commodity`)
	if err := p.Parse(); err != nil {
		t.Errorf(`tag-commodity failed: %v`, err)
	}
	if c, ok := p.Context().Commodities["USD"]; !ok {
		t.Errorf(`commodity did not create a USD commodity in the Context`)
	} else if len(c.GetTags()) != 2 {
		t.Errorf(`the commodity does not have two tags, it has %v`, len(c.GetTags()))
	} else if !c.HasTag("foo") {
		t.Errorf(`the commodity is not tagged with "foo"`)
	} else if !c.HasTag("bar") {
		t.Errorf(`the commodity is not tagged with "bar"`)
	}
	for _, tag := range []string{"foo", "bar"} {
		if tagged, ok := p.Context().Tags[tag]; !ok {
			t.Errorf(`the Context does not have a "%v" tag`, tag)
		} else if len(tagged) != 1 {
			t.Errorf(`the "%v" tag does not have exactly one object`, tag)
		} else if c, ok := tagged[0].(*core.Commodity); !ok {
			t.Errorf(`the object tagged with "%v" is not a Commodity`, tag)
		} else if c != p.Context().Commodities["USD"] {
			t.Errorf(`the object tagged with "%v" is not the USD commodity: %v`, tag, c.Name)
		}
	}
}

func TestTagCommodityFunction_ZeroOperands(t *testing.T) {
	p := createParser(`tag-commodity`)
	if p.Parse() == nil {
		t.Errorf(`tag-commodity succeeded with zero operands`)
	}
}

func TestTagCommodityFunction_NoTagOperands(t *testing.T) {
	p := createParser(`USD Dollar commodity USD tag-commodity`)
	if p.Parse() == nil {
		t.Errorf(`tag-commodity succeeded with no tag operands`)
	}
}

func TestTagCommodityFunction_NonStringCommodityNameOperand(t *testing.T) {
	p := createParser(`123 atoi foo tag-commodity`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`tag-commodity succeeded with a non-string commodity name operand`)
	}
}

func TestTagCommodityFunction_NonStringTagOperand(t *testing.T) {
	p := createParser(`USD Dollar commodity USD 123 atoi tag-commodity`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`tag-commodity succeeded with a non-string tag operand`)
	}
}

func TestTagCommodityFunction_NonexitentCommodity(t *testing.T) {
	p := createParser(`USD foo tag-commodity`)
	if p.Parse() == nil {
		t.Errorf(`tag-commodity succeeded with a nonexistent commodity`)
	}
}

func TestTagCommodityFunction_DuplicateTags(t *testing.T) {
	p := createParser(`USD Dollar commodity USD foo bar foo tag-commodity`)
	if err := p.Parse(); err != nil {
		t.Errorf(`tag-commodity failed: %v`, err)
	}
	if c, ok := p.Context().Commodities["USD"]; !ok {
		t.Errorf(`commodity did not create a USD commodity in the Context`)
	} else if len(c.GetTags()) != 2 {
		t.Errorf(`the commodity does not have two tags, it has %v`, len(c.GetTags()))
	} else if !c.HasTag("foo") {
		t.Errorf(`the commodity is not tagged with "foo"`)
	} else if !c.HasTag("bar") {
		t.Errorf(`the commodity is not tagged with "bar"`)
	}
	for _, tag := range []string{"foo", "bar"} {
		if tagged, ok := p.Context().Tags[tag]; !ok {
			t.Errorf(`the Context does not have a "%v" tag`, tag)
		} else if len(tagged) != 1 {
			t.Errorf(`the "%v" tag does not have exactly one object`, tag)
		} else if c, ok := tagged[0].(*core.Commodity); !ok {
			t.Errorf(`the object tagged with "%v" is not a Commodity`, tag)
		} else if c != p.Context().Commodities["USD"] {
			t.Errorf(`the object tagged with "%v" is not the USD commodity: %v`, tag, c.Name)
		}
	}
}

func TestTagCommodityFunction_TwoCommodities(t *testing.T) {
	p := createParser(`
		USD Dollar commodity
		JPY Yen commodity
		USD foo tag-commodity
		JPY foo tag-commodity`)
	if err := p.Parse(); err != nil {
		t.Errorf(`tag-commodity failed: %v`, err)
	}
	if tagged, ok := p.Context().Tags["foo"]; !ok {
		t.Errorf(`the Context does not have a "foo" tag`)
	} else if len(tagged) != 2 {
		t.Errorf(`the "foo" tag does not have two objects, it has %v`, len(tagged))
	} else {
		for _, cn := range []string{"USD", "JPY"} {
			if c, ok := p.Context().Commodities[cn]; !ok {
				t.Errorf(`commodity did not create a commodity named %v in the Context`, cn)
			} else if len(c.GetTags()) != 1 {
				t.Errorf(`the %v commodity does not have one tag, it has %v`, cn, len(c.GetTags()))
			} else if !c.HasTag("foo") {
				t.Errorf(`the %v commodity is not tagged with "foo"`, cn)
			} else {
				found := false
				for _, to := range tagged {
					if to == c {
						found = true
						break
					}
				}
				if !found {
					t.Errorf(`the %v commodity is not in Context.Tags["foo"]`, cn)
				}
			}
		}
	}
}

func TestUntagFunction(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account foo bar tag
		Assets:Account foo untag`)
	if err := p.Parse(); err != nil {
		t.Errorf(`untag failed: %v`, err)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf(`open did not create an account in the Context`)
	} else if len(a.GetTags()) != 1 {
		t.Errorf(`the account does not have 1 tag, it has %v`, len(a.GetTags()))
	} else if a.HasTag("foo") {
		t.Errorf(`the account is tagged with "foo"`)
	} else if !a.HasTag("bar") {
		t.Errorf(`the account is not tagged with "bar"`)
	} else if len(p.Context().Tags) != 1 {
		t.Errorf(`the Context has %v tags instead of 1`, len(p.Context().Tags))
	} else if tagged, ok := p.Context().Tags["bar"]; !ok {
		t.Errorf(`the Context does not have a "bar" tag`)
	} else if len(tagged) != 1 {
		t.Errorf(`the "bar" tag does not have exactly one object`)
	} else if a, ok := tagged[0].(*core.Account); !ok {
		t.Errorf(`the object tagged with "bar" is not an Account`)
	} else if a != p.Context().Accounts["Assets:Account"] {
		t.Errorf(`the object tagged with "bar" is not the Assets:Account account: %v`, a.Name)
	}
}

func TestUntagFunction_ZeroOperands(t *testing.T) {
	p := createParser(`untag`)
	if p.Parse() == nil {
		t.Errorf(`untag succeeded with zero operands`)
	}
}

func TestUntagFunction_NoTagOperands(t *testing.T) {
	p := createParser(`Assets:Account open Assets:Account untag`)
	if p.Parse() == nil {
		t.Errorf(`untag succeeded with no tag operands`)
	}
}

func TestUntagFunction_NonStringAccountNameOperand(t *testing.T) {
	p := createParser(`123 atoi foo untag`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`untag succeeded with a non-string account name operand`)
	}
}

func TestUntagFunction_NonStringTagOperand(t *testing.T) {
	p := createParser(`Assets:Account open Assets:Account 123 atoi untag`)
	p.Functions["atoi"] = atoi
	if p.Parse() == nil {
		t.Errorf(`untag succeeded with a non-string tag operand`)
	}
}

func TestUntagFunction_NonexitentAccount(t *testing.T) {
	p := createParser(`Assets:Account foo untag`)
	if p.Parse() == nil {
		t.Errorf(`untag succeeded with a nonexistent account`)
	}
}

func TestUntagFunction_ClosedAccount(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Account open
		Assets:Account close
		Assets:Account foo untag`)
	if p.Parse() == nil {
		t.Errorf(`untag succeeded with a closed account`)
	}
}

func TestUntagFunction_DuplicateTags(t *testing.T) {
	p := createParser(`
		Assets:Account open
		Assets:Account foo bar tag
		Assets:Account foo bar foo untag`)
	if err := p.Parse(); err != nil {
		t.Errorf(`untag failed: %v`, err)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf(`open did not create an account in the Context`)
	} else if len(a.GetTags()) != 0 {
		t.Errorf(`the account has %v tags instead of 0`, len(a.GetTags()))
	} else if len(p.Context().Tags) != 0 {
		t.Errorf(`the Context has a %v tags instead of 0`, len(p.Context().Tags))
	}
}

func TestUntagFunction_NonexistentTags(t *testing.T) {
	p := createParser(`
		Assets:Account open
		Assets:Account foo bar untag`)
	if err := p.Parse(); err != nil {
		t.Errorf(`untag failed: %v`, err)
	}
	if a, ok := p.Context().Accounts["Assets:Account"]; !ok {
		t.Errorf(`open did not create an account in the Context`)
	} else if len(a.GetTags()) != 0 {
		t.Errorf(`the account has %v tags instead of 0`, len(a.GetTags()))
	} else if len(p.Context().Tags) != 0 {
		t.Errorf(`the Context has a %v tags instead of 0`, len(p.Context().Tags))
	}
}

func TestUntagFunction_TwoAccounts(t *testing.T) {
	p := createParser(`
		2000 1 1 date
		Assets:Foo open
		Assets:Bar open
		Assets:Foo foo tag
		Assets:Bar foo tag
		Assets:Foo foo untag`)
	if err := p.Parse(); err != nil {
		t.Errorf(`untag failed: %v`, err)
	}
	if tagged, ok := p.Context().Tags["foo"]; !ok {
		t.Errorf(`the Context does not have a "foo" tag`)
	} else if len(tagged) != 1 {
		t.Errorf(`the "foo" tag does not have 1 object, it has %v`, len(tagged))
	} else if a, ok := p.Context().Accounts["Assets:Bar"]; !ok {
		t.Errorf(`open did not create an account named Assets:Bar in the Context`)
	} else if len(a.GetTags()) != 1 {
		t.Errorf(`Assets:Bar does not have 1 tag, it has %v`, len(a.GetTags()))
	} else if !a.HasTag("foo") {
		t.Errorf(`Assets:Bar is not tagged with "foo"`)
	} else if tagged[0] != a {
		t.Errorf(`Assets:Bar is not in Context.Tags["foo"]`)
	} else if a, ok := p.Context().Accounts["Assets:Foo"]; !ok {
		t.Errorf(`open did not create an account named Assets:Foo in the Context`)
	} else if len(a.GetTags()) != 0 {
		t.Errorf(`Assets:Foo has %v tags instead of 0`, len(a.GetTags()))
	}
}
