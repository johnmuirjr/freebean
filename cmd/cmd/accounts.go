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

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Print all accounts",
	Long: `The accounts subcommand reads a ledger from standard input
and prints all open accounts in CSV format.  The output includes a header.

The -c flag makes Freebean also print closed accounts.
The output will include a closing date column that specifies
the closing date; if it is empty, the account is open.

The -d flag specifies the date on which to stop parsing.
The date should be formatted "YYYY-MM-DD".  Parsing stops
at the end of the day, so accounts opened on that day
are included.  Freebean parses all input by default.

The -o flag makes Freebean print an additional column
that specifies the account's opening date.  If -c is also specified,
the opening date column will appear before the closing date column.`,
	Run: func(cmd *cobra.Command, args []string) {
		runAccounts()
	},
}

var accountsOptions = struct {
	Date                Date
	PrintClosedAccounts bool
	PrintOpeningDates   bool
}{}

func init() {
	rootCmd.AddCommand(accountsCmd)
	accountsCmd.Flags().VarP(&accountsOptions.Date, "date", "d", "date to stop parsing")
	accountsCmd.Flags().BoolVarP(&accountsOptions.PrintClosedAccounts, "print-closed-accounts", "c", false, "also print closed accounts")
	accountsCmd.Flags().BoolVarP(&accountsOptions.PrintOpeningDates, "print-opening-dates", "o", false, "also print opening dates")
}

func runAccounts() {
	done := &struct{}{}
	p := functions.NewParser(os.Stdin)
	p.AddCoreFunctions()
	date := core.Date(accountsOptions.Date)
	if !date.IsZero() {
		p.Functions["date"] = func(fn string, op parser.Operands, ctx *core.Context) error {
			if err := functions.DateFunction(fn, op, ctx); err != nil {
				return err
			} else if ctx.Date.After(date) {
				panic(done)
			}
			return nil
		}
	}
	defer func() {
		if r := recover(); r != nil && r != done {
			panic(r)
		}
		w := csv.NewWriter(os.Stdout)
		row := []string{"name"}
		if accountsOptions.PrintOpeningDates {
			row = append(row, "opening date")
		}
		if accountsOptions.PrintClosedAccounts {
			row = append(row, "closing date")
		}
		w.Write(row)
		for an, a := range p.Context().Accounts {
			if !accountsOptions.PrintClosedAccounts && a.IsClosed(p.Context().Date) {
				continue
			}
			row = append(row[:0], an)
			if accountsOptions.PrintOpeningDates {
				row = append(row, a.CreationDate.String())
			}
			if accountsOptions.PrintClosedAccounts {
				cd := ""
				if !a.ClosingDate.IsZero() {
					cd = a.ClosingDate.String()
				}
				row = append(row, cd)
			}
			w.Write(row)
		}
		w.Flush()
	}()
	if err := p.Parse(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
