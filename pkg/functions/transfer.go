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
	"strings"
)

type Transfer struct {
	Account      *core.Account
	LotName      string
	CreateLot    bool
	Quantity     core.Quantity
	ExchangeRate *core.ExchangeRate
	Comment      string
}

func (t Transfer) Lot(creationDate core.Date) *core.Lot {
	return &core.Lot{
		Name:         t.LotName,
		CreationDate: creationDate,
		Balance:      t.Quantity,
		ExchangeRate: t.ExchangeRate}
}

func (t Transfer) GetTransferQuantity() core.Quantity {
	if t.ExchangeRate != nil {
		return t.ExchangeRate.TotalPrice
	}
	return t.Quantity
}

func (t *Transfer) ExecuteTransfer(ctx *core.Context) error {
	if ctol, ok := t.Account.Lots[t.LotName]; !ok {
		if t.CreateLot {
			t.Account.Lots[t.LotName] = map[string]*core.Lot{t.Quantity.Commodity.Name: t.Lot(ctx.Date)}
		} else if len(t.LotName) == 0 {
			return fmt.Errorf(`account %v does not have a default lot`, t.Account.Name)
		} else {
			return fmt.Errorf(`account %v does not have a lot named "%v"`, t.Account.Name, t.LotName)
		}
	} else if l, ok := ctol[t.Quantity.Commodity.Name]; ok {
		l.Balance.Amount = l.Balance.Amount.Add(t.Quantity.Amount)
	} else {
		ctol[t.Quantity.Commodity.Name] = t.Lot(ctx.Date)
	}
	return nil
}

func ParseDecimal(q string) (decimal.Decimal, error) {
	return decimal.NewFromString(strings.ReplaceAll(q, ",", ""))
}

// Syntax: ACCOUNT AMOUNT COMMODITY -> Transfer
func ParseTransfer(op parser.Operands, ctx *core.Context) (*Transfer, error) {
	t := &Transfer{}
	if op.Length() < 3 {
		return t, fmt.Errorf("account name, quantity, and commodity name operands required, but too few given")
	}
	values := op.Pop(3)
	var an, q, cn string
	var c *core.Commodity
	var ok bool
	var e error
	if an, ok = values[0].(string); !ok {
		return t, fmt.Errorf("non-string account name: %v", values[0])
	} else if q, ok = values[1].(string); !ok {
		return t, fmt.Errorf("non-string quantity: %v", values[1])
	} else if cn, ok = values[2].(string); !ok {
		return t, fmt.Errorf("non-string commodity name: %v", values[2])
	} else if t.Quantity.Amount, e = ParseDecimal(q); e != nil {
		return t, fmt.Errorf("illegal decimal value %v: %v", q, e)
	}
	if t.Account, ok = ctx.Accounts[an]; !ok {
		return t, fmt.Errorf("nonexistent account: %v", an)
	} else if t.Account.IsClosed(ctx.Date) {
		return t, fmt.Errorf("closed account: %v", an)
	} else if c, ok = ctx.Commodities[cn]; !ok {
		return t, fmt.Errorf("nonexistent commodity: %v", cn)
	} else if len(t.Account.Commodities) != 0 {
		if _, ok = t.Account.Commodities[cn]; !ok {
			return t, fmt.Errorf("cannot transfer %v to or from account %v", cn, an)
		}
	}
	t.Quantity.Commodity = c
	return t, nil
}

// Syntax: ACCOUNT AMOUNT COMMODITY UNIT-AMOUNT UNIT-COMMODITY
// TOTAL-AMOUNT TOTAL-COMMODITY -> Transfer
func ParseTransferWithExchange(op parser.Operands, ctx *core.Context) (*Transfer, error) {
	t := &Transfer{ExchangeRate: &core.ExchangeRate{}}
	values := op.GetValues()
	for n := len(values) - 1; n >= 0; n-- {
		if _, ok := values[n].(string); !ok {
			values = values[n+1 : len(values)]
			break
		}
	}
	if len(values) < 7 {
		return t, fmt.Errorf("account name, quantity, commodity name, unit price amount, unit price commodity name, total price amount, and total price commodity name operands are required, but too few given")
	}
	values = op.Pop(7)
	var an, q, cn, upq, upcn, tpq, tpcn string
	var c *core.Commodity
	var ok bool
	var e error
	if an, ok = values[0].(string); !ok {
		return t, fmt.Errorf("non-string account name: %v", values[0])
	} else if q, ok = values[1].(string); !ok {
		return t, fmt.Errorf("non-string quantity: %v", values[1])
	} else if cn, ok = values[2].(string); !ok {
		return t, fmt.Errorf("non-string commodity name: %v", values[2])
	} else if t.Quantity.Amount, e = ParseDecimal(q); e != nil {
		return t, fmt.Errorf("illegal decimal value %v: %v", q, e)
	} else if upq, ok = values[3].(string); !ok {
		return t, fmt.Errorf("non-string unit price quantity: %v", values[3])
	} else if upcn, ok = values[4].(string); !ok {
		return t, fmt.Errorf("non-string unit price commodity name: %v", values[4])
	} else if t.ExchangeRate.UnitPrice.Amount, e = ParseDecimal(upq); e != nil {
		return t, fmt.Errorf("illegal decimal value %v: %v", upq, e)
	} else if tpq, ok = values[5].(string); !ok {
		return t, fmt.Errorf("non-string total price quantity: %v", values[5])
	} else if tpcn, ok = values[6].(string); !ok {
		return t, fmt.Errorf("non-string total price commodity name: %v", values[6])
	} else if t.ExchangeRate.TotalPrice.Amount, e = ParseDecimal(tpq); e != nil {
		return t, fmt.Errorf("illegal decimal value %v: %v", tpq, e)
	}
	if t.Account, ok = ctx.Accounts[an]; !ok {
		return t, fmt.Errorf("nonexistent account: %v", an)
	} else if t.Account.IsClosed(ctx.Date) {
		return t, fmt.Errorf("closed account: %v", an)
	}
	if c, ok = ctx.Commodities[cn]; !ok {
		return t, fmt.Errorf("nonexistent commodity: %v", cn)
	} else if len(t.Account.Commodities) != 0 {
		if _, ok = t.Account.Commodities[cn]; !ok {
			return t, fmt.Errorf("cannot transfer %v to or from account %v", cn, an)
		}
	}
	t.Quantity.Commodity = c
	if c, ok = ctx.Commodities[upcn]; !ok {
		return t, fmt.Errorf("nonexistent unit price commodity: %v", upcn)
	}
	t.ExchangeRate.UnitPrice.Commodity = c
	if c, ok = ctx.Commodities[tpcn]; !ok {
		return t, fmt.Errorf("nonexistent total price commodity: %v", tpcn)
	}
	t.ExchangeRate.TotalPrice.Commodity = c
	return t, nil
}
