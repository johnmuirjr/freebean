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

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Print all tags",
	Long: `The tags subcommand reads a ledger from standard input
and prints all tags in CSV format.  The output includes a header.

The -a flag makes Freebean print tagged accounts.  The output will include
a type column with the value "account" and a name column.  Note that this
flag makes the output repeat tags, once per tagged account.

The -c flag makes Freebean print tagged commodities.  The output will include
a type column with the value "commodity" and a name column.  Note that this
flag makes the output repeat tags, once per tagged commodity.

Specifying both -a and -c with interleave their results.

The -d flag specifies the date on which to stop parsing.
The date should be formatted "YYYY-MM-DD".  Parsing stops
at the end of the day, so accounts opened and commodities created
on that day are included.  Freebean parses all input by default.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTags()
	},
}

var tagsOptions = struct {
	Date             Date
	PrintAccounts    bool
	PrintCommodities bool
}{}

func init() {
	rootCmd.AddCommand(tagsCmd)
	tagsCmd.Flags().VarP(&tagsOptions.Date, "date", "d", "date to stop parsing")
	tagsCmd.Flags().BoolVarP(&tagsOptions.PrintAccounts, "print-accounts", "a", false, "print tagged accounts")
	tagsCmd.Flags().BoolVarP(&tagsOptions.PrintCommodities, "print-commodities", "c", false, "print tagged commodities")
}

func runTags() {
	done := &struct{}{}
	p := functions.NewParser(os.Stdin)
	p.AddCoreFunctions()
	date := core.Date(tagsOptions.Date)
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
		addlColumns := tagsOptions.PrintAccounts || tagsOptions.PrintCommodities
		if addlColumns {
			row = append(row, "type", "name")
		}
		w.Write(row)
		for tn, tagged := range p.Context().Tags {
			row = append(row[:0], tn)
			if addlColumns {
				for _, to := range tagged {
					switch v := to.(type) {
					case *core.Account:
						if tagsOptions.PrintAccounts && !v.IsClosed(p.Context().Date) {
							row = append(row[:1], "account", v.Name)
							w.Write(row)
						}
					case *core.Commodity:
						if tagsOptions.PrintCommodities {
							row = append(row[:1], "commodity", v.Name)
							w.Write(row)
						}
					}
				}
			} else {
				w.Write(row)
			}
		}
		w.Flush()
	}()
	if err := p.Parse(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
