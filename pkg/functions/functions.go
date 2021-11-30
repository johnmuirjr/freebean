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
	"strconv"
	"strings"
)

func GetCoreFunctions() map[string]Function {
	return map[string]Function{
		"add-notes":       AddNotesFunction,
		"assert":          AssertFunction,
		"assert-lot":      AssertLotFunction,
		"assert-lots-sum": AssertLotsSumFunction,
		"close":           CloseFunction,
		"close-lot":       CloseLotFunction,
		"comment":         CommentFunction,
		"commodity":       CommodityFunction,
		"create-lot":      CreateLotFunction,
		"date":            DateFunction,
		"lot":             LotFunction,
		"open":            OpenFunction,
		"set-comment":     SetCommentFunction,
		"tag":             TagFunction,
		"tag-commodity":   TagCommodityFunction,
		"untag":           UntagFunction,
		"xact":            XactFunction,     // TODO: test
		"xfer":            XferFunction,     // TODO: test
		"xfer-exch":       XferExchFunction, // TODO: test
	}
}

// AddNotesFunction adds notes to an account.
//
// Syntax: ACCOUNT (NOTE-NAME NOTE-VALUE)* add-notes ->
func AddNotesFunction(fn string, op parser.Operands, ctx *core.Context) error {
	values := op.GetValues()
	for n := len(values) - 1; n >= 0; n-- {
		if _, ok := values[n].(string); !ok {
			values = values[n+1 : len(values)]
			break
		}
	}
	if len(values) < 1 {
		return fmt.Errorf(`%v: account name operand required, but no operands given`, fn)
	} else if (len(values)-1)%2 != 0 {
		return fmt.Errorf(`%v: note name and note value operand pairs required, but odd number of operands given`, fn)
	}
	values = op.Pop(len(values))
	an := values[0].(string)
	if a, ok := ctx.Accounts[an]; !ok {
		return fmt.Errorf(`%v: nonexistent account: %v`, fn, an)
	} else if a.IsClosed(ctx.Date) {
		return fmt.Errorf(`%v: closed account: %v`, fn, an)
	} else {
		for n := 1; n < len(values); n += 2 {
			a.Notes[values[n].(string)] = values[n+1].(string)
		}
	}
	return nil
}

// AssertFunction asserts that the default lot within an account
// has the specified balance.
//
// Syntax: ACCOUNT AMOUNT COMMODITY assert ->
func AssertFunction(fn string, op parser.Operands, ctx *core.Context) error {
	if op.Length() < 3 {
		return fmt.Errorf(`%v: account name, amount, and commodity operands required, but too few given`, fn)
	}
	values := op.Pop(3)
	var an, as, cn string
	var q decimal.Decimal
	var e error
	var ok bool
	if an, ok = values[0].(string); !ok {
		return fmt.Errorf("%v: non-string account name: %v", fn, values[0])
	} else if as, ok = values[1].(string); !ok {
		return fmt.Errorf("%v: non-string quantity: %v", fn, values[1])
	} else if q, e = ParseDecimal(as); e != nil {
		return fmt.Errorf("%v: illegal decimal value %v: %v", fn, as, e)
	} else if cn, ok = values[2].(string); !ok {
		return fmt.Errorf("%v: non-string commodity name: %v", fn, values[2])
	}
	var acct *core.Account
	var lots map[string]*core.Lot
	var l *core.Lot
	if acct, ok = ctx.Accounts[an]; !ok {
		return fmt.Errorf("%v: nonexistent account: %v", fn, an)
	} else if acct.IsClosed(ctx.Date) {
		return fmt.Errorf("%v: closed account: %v", fn, an)
	} else if _, ok = ctx.Commodities[cn]; !ok {
		return fmt.Errorf("%v: nonexistent commodity: %v", fn, cn)
	} else if lots, ok = acct.Lots[""]; !ok {
		return fmt.Errorf("%v: account %v does not have a default lot", fn, an)
	} else if l, ok = lots[cn]; !ok {
		if !q.IsZero() {
			return fmt.Errorf("%v: default lot in account %v does not have %v", fn, an, cn)
		}
	} else if !l.Balance.Amount.Equal(q) {
		return fmt.Errorf("%v: default lot in account %v has %v, not asserted amount %v %v (difference of %v)", fn, an, l.Balance, q, l.Balance.Commodity, l.Balance.Amount.Sub(q))
	}
	return nil
}

// AssertLotFunction asserts that the specified lot within an account
// has the specified balance.
//
// Syntax: ACCOUNT LOT AMOUNT COMMODITY assert-lot ->
func AssertLotFunction(fn string, op parser.Operands, ctx *core.Context) error {
	if op.Length() < 4 {
		return fmt.Errorf(`%v: account name, lot name, amount, and commodity operands required, but too few given`, fn)
	}
	values := op.Pop(4)
	var an, ln, as, cn string
	var q decimal.Decimal
	var e error
	var ok bool
	if an, ok = values[0].(string); !ok {
		return fmt.Errorf("%v: non-string account name: %v", fn, values[0])
	} else if ln, ok = values[1].(string); !ok {
		return fmt.Errorf("%v: non-string lot name: %v", fn, values[1])
	} else if as, ok = values[2].(string); !ok {
		return fmt.Errorf("%v: non-string quantity: %v", fn, values[2])
	} else if q, e = ParseDecimal(as); e != nil {
		return fmt.Errorf("%v: illegal decimal value %v: %v", fn, as, e)
	} else if cn, ok = values[3].(string); !ok {
		return fmt.Errorf("%v: non-string commodity name: %v", fn, values[3])
	}
	var acct *core.Account
	var lots map[string]*core.Lot
	var l *core.Lot
	if acct, ok = ctx.Accounts[an]; !ok {
		return fmt.Errorf("%v: nonexistent account: %v", fn, an)
	} else if acct.IsClosed(ctx.Date) {
		return fmt.Errorf("%v: closed account: %v", fn, an)
	} else if _, ok = ctx.Commodities[cn]; !ok {
		return fmt.Errorf("%v: nonexistent commodity: %v", fn, cn)
	} else if lots, ok = acct.Lots[ln]; !ok {
		return fmt.Errorf(`%v: account %v does not have a lot named "%v"`, fn, an, ln)
	} else if l, ok = lots[cn]; !ok {
		if !q.IsZero() {
			return fmt.Errorf(`%v: lot "%v" in account %v does not have %v`, fn, ln, an, cn)
		}
	} else if !l.Balance.Amount.Equal(q) {
		return fmt.Errorf(`%v: lot "%v" in account %v has %v, not asserted amount %v %v (difference of %v)`, fn, ln, an, l.Balance, q, l.Balance.Commodity, l.Balance.Amount.Sub(q))
	}
	return nil
}

// AssertLotsSumFunction asserts that all of the lots in the specified account
// sum to the specified balance.
//
// Syntax: ACCOUNT AMOUNT COMMODITY assert-lots-sum ->
func AssertLotsSumFunction(fn string, op parser.Operands, ctx *core.Context) error {
	if op.Length() < 3 {
		return fmt.Errorf(`%v: account name, amount, and commodity operands required, but too few given`, fn)
	}
	values := op.Pop(3)
	var an, as, cn string
	var q decimal.Decimal
	var e error
	var ok bool
	if an, ok = values[0].(string); !ok {
		return fmt.Errorf("%v: non-string account name: %v", fn, values[0])
	} else if as, ok = values[1].(string); !ok {
		return fmt.Errorf("%v: non-string quantity: %v", fn, values[1])
	} else if q, e = ParseDecimal(as); e != nil {
		return fmt.Errorf("%v: illegal decimal value %v: %v", fn, as, e)
	} else if cn, ok = values[2].(string); !ok {
		return fmt.Errorf("%v: non-string commodity name: %v", fn, values[2])
	}
	var acct *core.Account
	if acct, ok = ctx.Accounts[an]; !ok {
		return fmt.Errorf("%v: nonexistent account: %v", fn, an)
	} else if acct.IsClosed(ctx.Date) {
		return fmt.Errorf("%v: closed account: %v", fn, an)
	} else if _, ok = ctx.Commodities[cn]; !ok {
		return fmt.Errorf("%v: nonexistent commodity: %v", fn, cn)
	} else {
		var sum decimal.Decimal
		for _, lmap := range acct.Lots {
			var l *core.Lot
			if l, ok = lmap[cn]; ok {
				sum = sum.Add(l.Balance.Amount)
			}
		}
		if !sum.Equal(q) {
			return fmt.Errorf(`%v: lots in account %v have a total of %v %v, not asserted amount %v %v (difference of %v)`, fn, an, sum, cn, q, cn, sum.Sub(q))
		}
	}
	return nil
}

// CloseFunction closes an account.
//
// Syntax: NAME close ->
func CloseFunction(fn string, op parser.Operands, ctx *core.Context) error {
	if op.Length() < 1 {
		return fmt.Errorf("%v: no operands given", fn)
	}
	values := op.Pop(1)
	var an string
	var ok bool
	if an, ok = values[0].(string); !ok {
		return fmt.Errorf("%v: non-string account name: %v", fn, values[0])
	}
	var acct *core.Account
	if acct, ok = ctx.Accounts[an]; !ok {
		return fmt.Errorf("%v: nonexistent account: %v", fn, an)
	} else if acct.IsClosed(ctx.Date) {
		return fmt.Errorf("%v: account is already closed: %v", fn, an)
	}
	for lotName, ctolots := range acct.Lots {
		if len(lotName) != 0 {
			for cn, lot := range ctolots {
				if !lot.Balance.Amount.IsZero() {
					return fmt.Errorf(`%v: cannot close account %v because lot "%v" has %v %v`, fn, an, lotName, lot.Balance.Amount, cn)
				}
			}
		}
	}
	acct.ClosingDate = ctx.Date
	return nil
}

// CloseLotFunction deletes a lot from an account.
//
// Syntax: ACCOUNT LOT close-lot ->
func CloseLotFunction(fn string, op parser.Operands, ctx *core.Context) error {
	if op.Length() < 2 {
		return fmt.Errorf("%v: account name and lot name operands required, but too few given", fn)
	}
	values := op.Pop(2)
	var an, ln string
	var ok bool
	if an, ok = values[0].(string); !ok {
		return fmt.Errorf("%v: non-string account name: %v", fn, values[0])
	} else if ln, ok = values[1].(string); !ok {
		return fmt.Errorf("%v: non-string lot name: %v", fn, values[1])
	}
	var acct *core.Account
	var lots map[string]*core.Lot
	if acct, ok = ctx.Accounts[an]; !ok {
		return fmt.Errorf("%v: nonexistent account: %v", fn, an)
	} else if acct.IsClosed(ctx.Date) {
		return fmt.Errorf("%v: closed account: %v", fn, an)
	} else if lots, ok = acct.Lots[ln]; !ok {
		return fmt.Errorf(`%v: nonexistent lot "%v" in account %v`, fn, ln, an)
	}
	for cn, lot := range lots {
		if !lot.Balance.Amount.IsZero() {
			return fmt.Errorf(`%v: cannot close lot "%v" in account %v because it has %v %v`, fn, ln, an, lot.Balance.Amount, cn)
		}
	}
	delete(acct.Lots, ln)
	return nil
}

// CommentFunction pops a string comment from the operand stack.
//
// Syntax: STRING comment ->
func CommentFunction(fn string, op parser.Operands, ctx *core.Context) error {
	if op.Length() < 1 {
		return fmt.Errorf("%v: no operands given", fn)
	}
	values := op.Pop(1)
	if _, ok := values[0].(string); !ok {
		return fmt.Errorf("%v: non-string operand: %v", fn, values[0])
	}
	return nil
}

// CommodityFunction creates a commodity.
//
// Syntax: NAME DESCRIPTION commodity ->
func CommodityFunction(fn string, op parser.Operands, ctx *core.Context) error {
	if op.Length() < 2 {
		return fmt.Errorf("%v: commodity name and description operands required, but too few given", fn)
	}
	values := op.Pop(2)
	var cn, d string
	var ok bool
	if cn, ok = values[0].(string); !ok {
		return fmt.Errorf("%v: non-string commodity name: %v", fn, values[0])
	} else if d, ok = values[1].(string); !ok {
		return fmt.Errorf("%v: non-string description: %v", fn, values[1])
	}
	if _, ok = ctx.Commodities[cn]; ok {
		return fmt.Errorf("%v: commodity already exists: %v", fn, cn)
	}
	ctx.Commodities[cn] = core.NewCommodity(cn, d, ctx.Date)
	return nil
}

// CreateLotFunction adds a lot name to a Transfer object on the operand stack.
// It asserts that the lot doesn't already exist or that it doesn't have
// the Transfer's commodity.
//
// Syntax: Transfer LOT create-lot -> Transfer
func CreateLotFunction(fn string, op parser.Operands, ctx *core.Context) error {
	if op.Length() < 2 {
		return fmt.Errorf("%v: transfer and lot name operands are required, but too few given", fn)
	}
	values := op.Pop(2)
	var t *Transfer
	var ln string
	var ok bool
	if t, ok = values[0].(*Transfer); !ok {
		return fmt.Errorf("%v: operand is not a transfer: %v", fn, values[0])
	} else if ln, ok = values[1].(string); !ok {
		return fmt.Errorf("%v: non-string lot name: %v", fn, values[1])
	}
	var ctolots map[string]*core.Lot
	if t.Account.IsClosed(ctx.Date) {
		return fmt.Errorf("%v: transfer refers to closed account: %v", fn, t.Account.Name)
	} else if ctolots, ok = t.Account.Lots[ln]; ok {
		if _, ok = ctolots[t.Quantity.Commodity.Name]; ok {
			return fmt.Errorf("%v: lot %v already contains %v", fn, ln, t.Quantity.Commodity.Name)
		}
	}
	t.LotName = ln
	t.CreateLot = true
	op.Push(t)
	return nil
}

// DateFunction sets the interpreter's current date.  It returns an error
// if the date jumps back in time.
//
// Syntax: YEAR MONTH DAY date ->
func DateFunction(fn string, op parser.Operands, ctx *core.Context) error {
	if op.Length() < 3 {
		return fmt.Errorf("%v: year, month, day operands required, but too few given", fn)
	}
	values := op.Pop(3)
	var ok bool
	var year, month, day string
	if year, ok = values[0].(string); !ok {
		return fmt.Errorf("%v: non-string year: %v", fn, values[0])
	} else if month, ok = values[1].(string); !ok {
		return fmt.Errorf("%v: non-string month: %v", fn, values[1])
	} else if day, ok = values[2].(string); !ok {
		return fmt.Errorf("%v: non-string day: %v", fn, values[2])
	}
	var y, m, dy int64
	var err error
	if y, err = strconv.ParseInt(year, 10, 32); err != nil {
		return fmt.Errorf("%v: illegal year %v: %v", fn, year, err)
	} else if m, err = strconv.ParseInt(month, 10, 32); err != nil {
		return fmt.Errorf("%v: illegal month %v: %v", fn, month, err)
	} else if dy, err = strconv.ParseInt(day, 10, 32); err != nil {
		return fmt.Errorf("%v: illegal day %v: %v", fn, day, err)
	}
	d := core.Date{int(y), int(m), int(dy)}
	if ctx.Date.After(d) {
		return fmt.Errorf("%v: specified date %v is before current date %v", fn, d, ctx.Date)
	}
	ctx.Date = d
	return nil
}

// LotFunction adds a lot name to a Transfer object on the operand stack.
// It asserts that the lot already exists.
//
// Syntax: Transfer LOT lot -> Transfer
func LotFunction(fn string, op parser.Operands, ctx *core.Context) error {
	if op.Length() < 2 {
		return fmt.Errorf("%v: transfer and lot name operands are required, but too few given", fn)
	}
	values := op.Pop(2)
	var t *Transfer
	var ln string
	var ok bool
	if t, ok = values[0].(*Transfer); !ok {
		return fmt.Errorf("%v: operand is not a transfer: %v", fn, values[0])
	} else if ln, ok = values[1].(string); !ok {
		return fmt.Errorf("%v: non-string lot name: %v", fn, values[1])
	} else if t.Account.IsClosed(ctx.Date) {
		return fmt.Errorf("%v: transfer refers to closed account: %v", fn, t.Account.Name)
	} else if _, ok = t.Account.Lots[ln]; !ok {
		return fmt.Errorf(`%v: account %v does not have a lot named "%v"`, fn, t.Account.Name, ln)
	}
	t.LotName = ln
	op.Push(t)
	return nil
}

// OpenFunction opens an account.  It returns an error if the specified account
// already exists and is open.
//
// Syntax: NAME COMMODITY* open ->
func OpenFunction(fn string, op parser.Operands, ctx *core.Context) error {
	values := op.GetValues()
	for n := len(values) - 1; n >= 0; n-- {
		if _, ok := values[n].(string); !ok {
			values = values[n+1 : len(values)]
			break
		}
	}
	if len(values) < 1 {
		return fmt.Errorf("%v: no operands given", fn)
	}
	values = op.Pop(len(values))
	an := values[0].(string)
	if !strings.HasPrefix(an, "Assets:") && !strings.HasPrefix(an, "Liabilities:") && !strings.HasPrefix(an, "Income:") && !strings.HasPrefix(an, "Expenses:") && !strings.HasPrefix(an, "Equity:") && an != "Equity" {
		return fmt.Errorf(`%v: account does not start with "Assets:", "Liabilities:", "Income:", "Expenses:", or "Equity:", and is not named "Equity": %v`, fn, an)
	}
	var acct *core.Account
	if acct, ok := ctx.Accounts[an]; ok {
		if !acct.IsClosed(ctx.Date) {
			return fmt.Errorf("%v: account already exists: %v", fn, an)
		}
	}
	acct = core.NewAccount(an, ctx.Date)
	for _, cn := range values[1:] {
		cname := cn.(string)
		if c, ok := ctx.Commodities[cname]; ok {
			acct.Commodities[cname] = c
		} else {
			return fmt.Errorf("%v: nonexistent commodity %v", fn, cname)
		}
	}
	ctx.Accounts[an] = acct
	return nil
}

// SetCommentFunction sets a Transfer's comment.
//
// Syntax: Transfer COMMENT set-comment -> Transfer
func SetCommentFunction(fn string, op parser.Operands, ctx *core.Context) error {
	if op.Length() < 2 {
		return fmt.Errorf(`%v: transfer and comment string operands required, but too few given`, fn)
	}
	values := op.Pop(2)
	if t, ok := values[0].(*Transfer); !ok {
		return fmt.Errorf("%v: not a transfer: %v", fn, values[0])
	} else if comment, ok := values[1].(string); !ok {
		return fmt.Errorf("%v: non-string comment: %v", fn, values[1])
	} else {
		t.Comment = comment
		op.Push(t)
	}
	return nil
}

// TagFunction tags an account.
//
// Syntax: ACCOUNT TAG+ tag ->
func TagFunction(fn string, op parser.Operands, ctx *core.Context) error {
	values := op.GetValues()
	for n := len(values) - 1; n >= 0; n-- {
		if _, ok := values[n].(string); !ok {
			values = values[n+1 : len(values)]
			break
		}
	}
	if len(values) < 2 {
		return fmt.Errorf("%v: account name and at least one tag operand required, but too few operands given", fn)
	}
	values = op.Pop(len(values))
	an := values[0].(string)
	var acct *core.Account
	var ok bool
	if acct, ok = ctx.Accounts[an]; !ok {
		return fmt.Errorf("%v: tagging nonexistent account: %v", fn, an)
	} else if acct.IsClosed(ctx.Date) {
		return fmt.Errorf("%v: closed account: %v", fn, an)
	}
	for _, t := range values[1:] {
		tag := t.(string)
		if tts, ok := ctx.Tags[tag]; ok {
			found := false
			for _, tagged := range tts {
				if tagged == acct {
					found = true
					break
				}
			}
			if !found {
				ctx.Tags[tag] = append(tts, acct)
			}
		} else {
			ctx.Tags[tag] = []core.TagTarget{acct}
		}
		acct.AddTag(tag)
	}
	return nil
}

// TagCommodityFunction tags a commodity.
//
// Syntax: COMMODITY TAG+ tag-commodity ->
func TagCommodityFunction(fn string, op parser.Operands, ctx *core.Context) error {
	values := op.GetValues()
	for n := len(values) - 1; n >= 0; n-- {
		if _, ok := values[n].(string); !ok {
			values = values[n+1 : len(values)]
			break
		}
	}
	if len(values) < 2 {
		return fmt.Errorf("%v: commodity name and at least one tag operand required, but too few operands given", fn)
	}
	values = op.Pop(len(values))
	cn := values[0].(string)
	var c *core.Commodity
	var ok bool
	if c, ok = ctx.Commodities[cn]; !ok {
		return fmt.Errorf("%v: tagging nonexistent commodity: %v", fn, cn)
	}
	for _, t := range values[1:] {
		tag := t.(string)
		if tts, ok := ctx.Tags[tag]; ok {
			found := false
			for _, tagged := range tts {
				if tagged == c {
					found = true
					break
				}
			}
			if !found {
				ctx.Tags[tag] = append(tts, c)
			}
		} else {
			ctx.Tags[tag] = []core.TagTarget{c}
		}
		c.AddTag(tag)
	}
	return nil
}

// UntagFunction untags an account.
//
// Syntax: ACCOUNT TAG+ untag ->
func UntagFunction(fn string, op parser.Operands, ctx *core.Context) error {
	values := op.GetValues()
	for n := len(values) - 1; n >= 0; n-- {
		if _, ok := values[n].(string); !ok {
			values = values[n+1 : len(values)]
			break
		}
	}
	if len(values) < 2 {
		return fmt.Errorf("%v: account name and at least one tag operand required, but too few operands given", fn)
	}
	values = op.Pop(len(values))
	an := values[0].(string)
	if a, ok := ctx.Accounts[an]; !ok {
		return fmt.Errorf("%v: tagging nonexistent account: %v", fn, an)
	} else if a.IsClosed(ctx.Date) {
		return fmt.Errorf("%v: closed account: %v", fn, an)
	} else {
		for _, t := range values[1:] {
			tag := t.(string)
			if tts, ok := ctx.Tags[tag]; ok {
				n := len(tts)
				for m := 0; m < n; {
					if tts[m] == a {
						n--
						tts[m] = tts[n]
					} else {
						m++
					}
				}
				tts = tts[:n]
				if len(tts) != 0 {
					ctx.Tags[tag] = tts
				} else {
					delete(ctx.Tags, tag)
				}
			}
			a.RemoveTag(tag)
		}
	}
	return nil
}

// XactFunction effects a series of transfers.
//
// Syntax: ENTITY DESCRIPTION Transfer+ (NOTE-NAME NOTE-VALUE)* xact ->
func XactFunction(fn string, op parser.Operands, ctx *core.Context) error {
	t, err := ParseTransaction(op, ctx)
	if err == nil {
		if err = t.Execute(ctx); err != nil {
			err = fmt.Errorf("%v: %v", fn, err)
		}
	} else {
		err = fmt.Errorf("%v: %v", fn, err)
	}
	return err
}

// XferFunction pushes a Transfer object onto the operand stack.
// It does not create an exchange rate and it targets the default lot.
//
// Syntax: ACCOUNT AMOUNT COMMODITY xfer -> Transfer
func XferFunction(fn string, op parser.Operands, ctx *core.Context) error {
	t, err := ParseTransfer(op, ctx)
	if err == nil {
		op.Push(t)
	} else {
		err = fmt.Errorf("%v: %v", fn, err)
	}
	return err
}

// XferExchFunction pushes a Transfer object onto the operand stack with an
// exchange rate.
//
// Syntax: ACCOUNT AMOUNT COMMODITY UNIT-AMOUNT UNIT-COMMODITY
// TOTAL-AMOUNT TOTAL-COMMODITY xfer-exch -> Transfer
func XferExchFunction(fn string, op parser.Operands, ctx *core.Context) error {
	t, err := ParseTransferWithExchange(op, ctx)
	if err == nil {
		op.Push(t)
	} else {
		err = fmt.Errorf("%v: %v", fn, err)
	}
	return err
}
