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

var lotsCmd = &cobra.Command{
	Use:   "lots",
	Short: "Print all lots",
	Long: `The lots subcommand reads a ledger from standard input
and prints all lots in all open accounts in CSV format.  The output
includes a header.  Lots without exchange rates have blank unit price
and total price columns.

The -a flag makes Freebean print lot assertions in the ledger language
instead of CSV.

The -d flag specifies the date on which to stop parsing.
The date should be formatted "YYYY-MM-DD".  Parsing stops
at the end of the day, so accounts opened and lots created
on that day are included.  Freebean parses all input by default.

The -D flag makes Freebean also print default (unnamed) lots.
Default lots have blank lot names.`,
	Run: func(cmd *cobra.Command, args []string) {
		runLots()
	},
}

var lotsOptions = struct {
	Date             Date
	PrintDefaultLots bool
	PrintAssertions  bool
}{}

func init() {
	rootCmd.AddCommand(lotsCmd)
	lotsCmd.Flags().BoolVarP(&lotsOptions.PrintDefaultLots, "print-default-lots", "D", false, "also print default lots")
	lotsCmd.Flags().VarP(&lotsOptions.Date, "date", "d", "date to stop parsing")
	lotsCmd.Flags().BoolVarP(&lotsOptions.PrintAssertions, "print-assertions", "a", false, "print assertions instead of CSV")
}

func runLots() {
	done := &struct{}{}
	p := functions.NewParser(os.Stdin)
	p.AddCoreFunctions()
	date := core.Date(lotsOptions.Date)
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
		row := []string{"account name", "lot name", "commodity", "balance", "unit price", "total price"}
		printRow := func(vals []string) { w.Write(row) }
		if lotsOptions.PrintAssertions {
			printRow = func(vals []string) {
				if len(vals[1]) == 0 {
					fmt.Printf("%v %v assert\n", vals[0], vals[3])
				} else {
					fmt.Printf("%v %v %v assert-lot\n", vals[0], vals[1], vals[3])
				}
			}
		} else {
			w.Write(row)
		}
		for an, a := range p.Context().Accounts {
			if !a.IsClosed(p.Context().Date) {
				row = append(row[:0], an)
				for ln, ctol := range a.Lots {
					if !lotsOptions.PrintDefaultLots && len(ln) == 0 {
						continue
					}
					row = append(row[:1], ln)
					for cn, l := range ctol {
						row = append(row[:2], cn, l.Balance.String())
						if l.ExchangeRate != nil {
							row = append(row, l.ExchangeRate.UnitPrice.String(), l.ExchangeRate.TotalPrice.String())
						} else {
							row = append(row, "", "")
						}
						printRow(row)
					}
				}
			}
		}
		w.Flush()
	}()
	if err := p.Parse(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
