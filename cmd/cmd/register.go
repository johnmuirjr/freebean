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

package cmd

import (
	"encoding/csv"
	"fmt"
	"github.com/jtvaughan/freebean/pkg/core"
	"github.com/jtvaughan/freebean/pkg/functions"
	"github.com/jtvaughan/freebean/pkg/parser"
	"github.com/spf13/cobra"
	"os"
)

var registerCmd = &cobra.Command{
	Use:   "register [account] [commodity]",
	Short: "Print all transfers affecting an account",
	Long: `The register subcommand reads a ledger from standard input
and prints all transfers affecting the specified account in
the specified commodity in CSV format.  The output includes a header
with each transfer's date, its transaction's entity name,
the amount transferred, and the current balance.

The -s flag specifies the date on which to start printing transfers.
The date should be formatted "YYYY-MM-DD".  Freebean parses all input
by default.

The -e flag specifies the date on which to stop parsing.
The date should be formatted "YYYY-MM-DD".  Parsing stops
at the end of the day, so transfers affecting the account
on that day are included.  Freebean parses all input by default.

The -l flag makes Freebean limit results to the specified lot
within the account.  Freebean limits its results to the default
lot by default.

The -n flag makes Freebean also print the specified note
attached to each transfer's transaction.  This adds a column
with the note's name and value to the output.  If the note
is absent from a transfer's transaction, the column value
will be blank.  The -n flag may be repeated any number of times.

The -x flag makes Freebean also print exchange rates.
This adds unit price and total price columns to the output.
Transfers without exchange rates will have blank values
in these columns.

The -z flag makes Freebean start the account with a zero balance
on the start date specified by the -s flag.  Freebean uses the
account's real balance by default regardless of the start date.
This flag only makes sense when combined with -s.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runRegister(args[0], args[1])
	},
}

var registerOptions = struct {
	StartDate            Date
	EndDate              Date
	LotName              string
	PrintExchangeRates   bool
	StartWithZeroBalance bool
	Notes                []string
}{}

func init() {
	rootCmd.AddCommand(registerCmd)
	registerCmd.Flags().VarP(&registerOptions.StartDate, "start-date", "s", "date to start printing transfers")
	registerCmd.Flags().VarP(&registerOptions.EndDate, "end-date", "e", "date to stop printing transfers")
	registerCmd.Flags().StringVarP(&registerOptions.LotName, "lot", "l", "", "limit results to this lot")
	registerCmd.Flags().BoolVarP(&registerOptions.PrintExchangeRates, "print-exchange-rates", "x", false, "also print exchange rates")
	registerCmd.Flags().BoolVarP(&registerOptions.StartWithZeroBalance, "zero-balance", "z", false, "start with a zero balance")
	registerCmd.Flags().StringSliceVarP(&registerOptions.Notes, "note", "n", nil, "also print these transaction notes")
}

func runRegister(accountName, commodityName string) {
	done := &struct{}{}
	p := functions.NewParser(os.Stdin)
	p.AddCoreFunctions()

	w := csv.NewWriter(os.Stdout)
	row := []string{"date", "entity", "amount", "balance"}
	if registerOptions.PrintExchangeRates {
		row = append(row, "unit price", "total price")
	}
	row = append(row, registerOptions.Notes...)
	w.Write(row)

	var balance *core.Quantity
	if registerOptions.StartWithZeroBalance {
		balance = &core.Quantity{Commodity: &core.Commodity{Name: commodityName}}
	}
	startDate := core.Date(registerOptions.StartDate)
	endDate := core.Date(registerOptions.EndDate)
	if !endDate.IsZero() {
		p.Functions["date"] = func(fn string, op parser.Operands, ctx *core.Context) error {
			if err := functions.DateFunction(fn, op, ctx); err != nil {
				return err
			} else if ctx.Date.After(endDate) {
				panic(done)
			}
			return nil
		}
	}
	p.Functions["xact"] = func(fn string, op parser.Operands, ctx *core.Context) error {
		var xact functions.Transaction
		var err error
		if xact, err = functions.ParseTransaction(op, ctx); err != nil {
			return err
		} else if err = xact.Execute(ctx); err != nil {
			return err
		}
		if ctx.Date.EqualOrAfter(startDate) {
			for _, t := range xact.Transfers {
				if t.Account.Name == accountName && t.LotName == registerOptions.LotName && t.Quantity.Commodity.Name == commodityName {
					row = append(row[:0], ctx.Date.String(), xact.Entity, t.Quantity.String())
					if balance != nil {
						balance.Amount = balance.Amount.Add(t.Quantity.Amount)
						row = append(row, balance.String())
					} else {
						row = append(row, t.Account.Lots[t.LotName][commodityName].Balance.String())
					}
					if registerOptions.PrintExchangeRates {
						if t.ExchangeRate != nil {
							row = append(row, t.ExchangeRate.UnitPrice.String(), t.ExchangeRate.TotalPrice.String())
						} else {
							row = append(row, "", "")
						}
					}
					for _, n := range registerOptions.Notes {
						row = append(row, xact.Notes[n])
					}
					w.Write(row)
				}
			}
		}
		return nil
	}
	defer func() {
		if r := recover(); r != nil && r != done {
			panic(r)
		}
		w.Flush()
	}()
	if err := p.Parse(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
