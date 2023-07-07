// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gstrings is a more capable, UTF-8 aware version of the standard strings utility.
package hexdump

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func NewScanConfig(out io.Writer) *ScanConfig {
	return &ScanConfig{
		Min:     6,
		Max:     256,
		Ascii:   false,
		Tab:     false,
		Search:  "",
		Most:    0,
		Offset:  false,
		Verbose: false,
		Out:     out,
	}
}

type ScanConfig struct {
	Out     io.Writer
	Search  string //  search ASCII sub-string
	Min     int    // minimum length of UTF-8 strings printed, in runes
	Max     int    // maximum length of UTF-8 strings printed, in runes
	Most    int    //  print at most n places
	Ascii   bool   // restrict strings to ASCII
	Tab     bool   // print strings separated by tabs other than new lines
	Offset  bool   //  show file name and offset of start of each string
	Verbose bool   // display all input data.  Without the -v option, any output lines, which would be identical to the immediately preceding output line(except for the input offsets), are replaced with a line comprised of a single asterisk.
}

type Scanner struct {
	*ScanConfig
	file string

	lastPrint string

	printable  []rune
	pos        int64
	printTimes int
}

func (c *ScanConfig) NewScanner(file string) *Scanner {
	return &Scanner{
		file:       file,
		printable:  make([]rune, 0, c.Max),
		ScanConfig: c,
	}
}

func If[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func (f *Scanner) Scan(in io.RuneReader) error {
	var r rune
	var wid int
	var err error

	f.Min = If(f.Min <= 0, 6, f.Min)
	f.Max = If(f.Max <= 0, 256, f.Max)

	// One string per loop.
	for ; ; f.pos += int64(wid) {
		if r, wid, err = in.ReadRune(); err != nil {
			return err
		}
		if !strconv.IsPrint(r) || f.Ascii && r >= 0xFF {
			f.print()
			continue
		}
		// It's printable. Keep it.
		f.printable = append(f.printable, r)
		if len(f.printable) >= cap(f.printable) {
			f.print()
		}
	}
}

func (f *Scanner) print() {
	if len(f.printable) < f.Min {
		f.printable = f.printable[:0]
		return
	}

	s := string(f.printable)
	if f.Search == "" || strings.Contains(s, f.Search) {
		if !f.Verbose {
			if f.lastPrint == s {
				s = "*"
			} else {
				f.lastPrint = s
			}
		}
		if f.Offset {
			s = fmt.Sprintf("%s:#%d:\t%s", f.file, f.pos-int64(len(s)), s)
		}

		if f.Tab {
			fmt.Fprint(f.Out, s)
			fmt.Fprint(f.Out, "\t")
		} else {
			fmt.Fprintln(f.Out, s)
		}
		f.printTimes++

		if f.Most > 0 && f.printTimes >= f.Max {
			fmt.Fprintln(f.Out)
			os.Exit(0)
		}
	}

	f.printable = f.printable[:0]
}
