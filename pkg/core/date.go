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

import (
	"fmt"
	"time"
)

type Date struct {
	Year  int
	Month int
	Day   int
}

func FromTime(t time.Time) Date {
	return Date{t.Year(), int(t.Month()), t.Day()}
}

func ParseDate(s string) (Date, error) {
	if t, err := time.Parse("2006-01-02", s); err != nil {
		return Date{}, err
	} else {
		return FromTime(t), nil
	}
}

func (d Date) ToTime() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
}

func (d Date) After(u Date) bool {
	return d.ToTime().After(u.ToTime())
}

func (d Date) Before(u Date) bool {
	return d.ToTime().Before(u.ToTime())
}

func (d Date) BeforeOrEqual(u Date) bool {
	return d.Before(u) || d.Equal(u)
}

func (d Date) Equal(u Date) bool {
	return d.Year == u.Year && d.Month == u.Month && d.Day == u.Day
}

func (d Date) EqualOrAfter(u Date) bool {
	return d.Equal(u) || d.After(u)
}

func (d Date) IsZero() bool { return d.Equal(Date{}) }

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}
