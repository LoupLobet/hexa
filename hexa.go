package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"unicode"

	"9fans.net/go/acme"
)

type WindowLine struct {
	bytes  []byte
	nbytes int
	addr   int
}

type Window struct {
	name        string
	acme        *acme.Win
	addrLen     int
	lines       []*WindowLine
}

var (
	BytePerLine = flag.Int("c", 16, "byte per line")
)

func main() {
	flag.Parse()
	for _, v := range flag.Args() {
		w, err := newWindow(v)
		if err != nil {
			w.acme.Del(true)
			log.Fatal(err)
		}
	}
}

func newWindow(path string) (*Window, error) {
	win, err := acme.New()
	if err != nil {
		return nil, err
	}
	w := &Window{
		acme: win,
		addrLen: 8,
		lines: []*WindowLine{},
	}
	if err := w.loadFile(path); err != nil {
		return nil, err
	}
	// render all lines for the first time
	offset := 0
	for _, line := range w.lines {
		n, err := w.printLine(line, []byte(fmt.Sprintf("#%d", offset)))
		if err != nil {
			return nil, err
		}
		offset += n
	}
	// window name
	if err := w.acme.Ctl("name %s/hexa", path); err != nil {
		return nil, err
	}
	// mark window as clean and go move cursor to first line
	if err := win.Ctl("clean"); err != nil {
		return nil, err
	}
	if _, err := win.Write("addr", []byte{'0'}); err != nil {
		return nil, err
	}
	if err := win.Ctl("dot=addr"); err != nil {
		return nil, err
	}
	if err := win.Ctl("show"); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Window) loadFile(path string) error {
	// read file bytePerLine bytes at a time
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	for addr := 0; ; addr += *BytePerLine {
		b := make([]byte, *BytePerLine)
		n, err := f.Read(b)
		line := &WindowLine{
			addr: addr,
			nbytes: n,
			bytes: b,
		}
		w.lines = append(w.lines, line)
		if n < len(b) || err != nil {
			return err
		}
	}
}

func (w *Window) printLine(l *WindowLine, acmeAddr []byte) (int, error) {
	var b []byte
	// addr
	b = append(b, []byte(fmt.Sprintf("%08x", l.addr))...)
	b = append(b, ' ')
	// bytes
	strByte := make([]byte, 2)
	for i, v := range l.bytes {
		if v == 0 {
			strByte[0] = ' '
			strByte[1] = ' '
		} else {
			hex.Encode(strByte, []byte{v})
		}
		if i % 4 == 0 {
			b = append(b, ' ')
		}
		b = append(b, strByte...)
		b = append(b, ' ')
	}
	b = append(b, ' ')
	// ascii
	for _, v := range l.bytes[:l.nbytes] {
		if unicode.IsPrint(rune(v)) {
			b = append(b, v)
		} else {
			b = append(b, '.')
		}
	}
	b = append(b, '\n')
	// move cursor to the appropriate position
	if _, err := w.acme.Write("addr", acmeAddr); err != nil {
		return 0, err
	}
	if err := w.acme.Ctl("dot=addr"); err != nil {
		return 0, err
	}
	if _, err := w.acme.Write("data", b); err != nil {
		return 0, err
	}
	return l.len(), nil
}


func (l *WindowLine) len() int {
	return 8 + 2 + (*BytePerLine * 3) + (*BytePerLine / 4) + *BytePerLine + 1
}
