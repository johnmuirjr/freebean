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

package core

type Account struct {
	Name         string
	CreationDate Date
	ClosingDate  Date
	Commodities  map[string]*Commodity
	Lots         map[string]map[string]*Lot // lot name -> commodity name -> *Lot
	Tags         map[string]bool
	Notes        map[string]string
}

func NewAccount(name string, creationDate Date) *Account {
	return &Account{
		Name:         name,
		CreationDate: creationDate,
		Commodities:  map[string]*Commodity{},
		Lots:         map[string]map[string]*Lot{"": map[string]*Lot{}},
		Tags:         map[string]bool{},
		Notes:        map[string]string{}}
}

func (a *Account) IsClosed(date Date) bool {
	return !a.ClosingDate.Equal(Date{}) && date.EqualOrAfter(a.ClosingDate)
}

func (a *Account) AddTag(tag string) {
	a.Tags[tag] = true
}

func (a *Account) GetTags() []string {
	tags := make([]string, len(a.Tags))[:0]
	for tag, _ := range a.Tags {
		tags = append(tags, tag)
	}
	return tags
}

func (a *Account) HasTag(tag string) bool {
	_, ok := a.Tags[tag]
	return ok
}

func (a *Account) RemoveTag(tag string) {
	delete(a.Tags, tag)
}
