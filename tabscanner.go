package struct2ts

import "io"

type tabScanner struct {
	output io.Writer
	tabs   []byte
}

func (ts *tabScanner) Write(bs []byte) (n int, err error) {
	p := 0
	var rn int
	for pos, b := range bs {
		switch b {
		case '\n':
			rn, err = ts.output.Write(bs[p : pos+1])
			n += rn
			if err != nil {
				return
			}
			_, err = ts.output.Write(ts.tabs)
			if err != nil {
				return
			}
			p = pos + 1
		}
	}
	rn, err = ts.output.Write(bs[p:])
	return n + rn, err
}

func newTabScanner(w io.Writer, tabs string) io.Writer {
	return &tabScanner{
		output: w,
		tabs:   []byte(tabs),
	}
}
