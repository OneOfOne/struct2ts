package struct2ts

import (
	"fmt"
	"io"
	"reflect"
)

var zeroValues = map[string]string{
	"number":  "0",
	"boolean": "false",
	"string":  "''",
	"object":  "{}",
	"map":     "{}",
	"array":   "[]",
}

type Struct struct {
	Name   string
	Fields []*Field

	t reflect.Type
}

func (s *Struct) RenderTo(opts *Options, w io.Writer) (err error) {
	if _, err = fmt.Fprintf(w, "// struct2ts:%s.%s\n", s.t.PkgPath(), s.Name); err != nil {
		return
	}

	if opts.InterfaceOnly {
		if opts.ES6 { // no interfaces in js
			return
		}
		_, err = fmt.Fprintf(w, "interface %s {\n", s.Name)
	} else {
		_, err = fmt.Fprintf(w, "class %s {\n", s.Name)
	}

	if err != nil {
		return
	}

	if !opts.ES6 {
		if err = s.RenderFields(opts, w); err != nil {
			return
		}
	}

	if err = s.RenderConstructor(opts, w); err != nil {
		return
	}

	if err = s.RenderToObject(opts, w); err != nil {
		return
	}

	_, err = fmt.Fprint(w, "}")
	return
}

func (s *Struct) RenderFields(opts *Options, w io.Writer) (err error) {
	for _, f := range s.Fields {
		if err = f.RenderTopLevel(w, opts); err != nil {
			return
		}
	}

	return
}

func (s *Struct) RenderConstructor(opts *Options, w io.Writer) (err error) {
	if opts.NoConstructor || opts.InterfaceOnly {
		return
	}

	if opts.ES6 {
		fmt.Fprintf(w, "%sconstructor(data = null) {\n", opts.indents[1])
		fmt.Fprintf(w, "%sconst d = (data && typeof data === 'object') ? ToObject(data) : {};\n", opts.indents[2])
	} else {
		fmt.Fprintf(w, "\n%sconstructor(data?: any) {\n", opts.indents[1])
		fmt.Fprintf(w, "%sconst d: any = (data && typeof data === 'object') ? ToObject(data) : {};\n", opts.indents[2])
	}

	for _, f := range s.Fields {
		if err = f.RenderCtor(w, opts); err != nil {
			return
		}
	}

	_, err = fmt.Fprintf(w, "%s}\n", opts.indents[1])

	return
}

func (s *Struct) RenderToObject(opts *Options, w io.Writer) (err error) {
	if opts.NoToObject || opts.InterfaceOnly {
		return
	}

	if opts.ES6 {
		fmt.Fprintf(w, "\n%stoObject() {\n", opts.indents[1])
		fmt.Fprintf(w, "%sconst cfg = {};\n", opts.indents[2])
	} else {
		fmt.Fprintf(w, "\n%stoObject(): any {\n", opts.indents[1])
		fmt.Fprintf(w, "%sconst cfg: any = {};\n", opts.indents[2])
	}

	for _, f := range s.Fields {
		t := f.Type(opts, true)
		switch {
		case t == "Date" && f.TsType != "number":
			fmt.Fprintf(w, "%scfg.%s = 'string';\n", opts.indents[2], f.Name)
		case t == "number":
			fmt.Fprintf(w, "%scfg.%s = 'number';\n", opts.indents[2], f.Name)
		}
	}
	_, err = fmt.Fprintf(w, "%sreturn ToObject(this, cfg);\n%s}\n", opts.indents[2], opts.indents[1])
	return
}
