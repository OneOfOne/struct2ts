//go:generate sh ./cmd/genHelpers.sh helpers.ts helpers_gen.go

package struct2ts

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
)

type Options struct {
	Indent string

	NoAssignDefaults bool
	InterfaceOnly    bool

	MarkOptional  bool
	NoConstructor bool
	NoToObject    bool
	NoExports     bool
	NoHelpers     bool
	NoDate        bool
	ES6           bool

	indents [3]string
}

func New(opts *Options) *StructToTS {
	if opts == nil {
		opts = &Options{}
	}

	if opts.Indent == "" {
		opts.Indent = "\t"
	}

	for i := range opts.indents {
		opts.indents[i] = strings.Repeat(opts.Indent, i)
	}

	return &StructToTS{
		seen: map[reflect.Type]*Struct{},
		opts: opts,
	}
}

type StructToTS struct {
	structs []*Struct
	seen    map[reflect.Type]*Struct
	opts    *Options
}

func (s *StructToTS) Add(v interface{}) *Struct { return s.AddWithName(v, "") }

func (s *StructToTS) AddWithName(v interface{}, name string) *Struct {
	var t reflect.Type
	switch v := v.(type) {
	case reflect.Type:
		t = v
	case reflect.Value:
		t = v.Type()
	default:
		t = reflect.TypeOf(v)
	}

	return s.addType(t, name, "")
}

func (s *StructToTS) addType(t reflect.Type, name, prefix string) (out *Struct) {
	t = indirect(t)

	if out = s.seen[t]; out != nil {
		return out
	}

	if name == "" {
		name = t.Name()
	}

	out = &Struct{
		Name:   prefix + name,
		Fields: make([]*Field, 0, t.NumField()),

		t: t,
	}

	s.seen[t] = out

	for i := 0; i < t.NumField(); i++ {
		var (
			sf  = t.Field(i)
			sft = sf.Type
			tf  Field
			k   = sft.Kind()
		)

		if k == reflect.Ptr {
			tf.CanBeNull = true
			sft = indirect(sft)
			k = sft.Kind()
		}

		if tf.setProps(sf, sft) {
			continue
		}

		switch {
		case k == reflect.Map:
			tf.TsType, tf.KeyType, tf.ValType = "map", stripType(sft.Key()), stripType(sft.Elem())

			if isStruct(sft.Elem()) {
				tf.ValType = s.addType(sft.Elem(), "", out.Name).Name
			}

		case k == reflect.Slice, k == reflect.Array:
			tf.TsType, tf.ValType = "array", stripType(sft.Elem())

			if isStruct(sft.Elem()) {
				tf.ValType = s.addType(sft.Elem(), "", out.Name).Name
			}

		case k == reflect.Struct:
			if isDate(sft) {
				break
			}
			tf.TsType = "object"
			tf.ValType = s.addType(sft, "", out.Name).Name

		case k == reflect.Interface:
			tf.TsType, tf.ValType = "object", ""

		case tf.TsType != "": // native type
		default:
			log.Println("unhandled", k, sft)
		}

		out.Fields = append(out.Fields, &tf)
	}

	s.structs = append(s.structs, out)
	return
}

func (s *StructToTS) RenderTo(w io.Writer) (err error) {
	buf := bufio.NewWriter(w)
	defer buf.Flush()

	if s.opts.ES6 {
		io.WriteString(w, "'use strict';\n")
	}

	if !s.opts.NoHelpers {
		io.WriteString(w, "\n// helpers")
		if s.opts.ES6 {
			fmt.Fprint(w, es6_helpers)
		} else {
			fmt.Fprint(w, ts_helpers)
		}
		io.WriteString(w, "\n")
	}

	io.WriteString(w, "// classes\n")
	for _, st := range s.structs {
		if err = st.RenderTo(s.opts, buf); err != nil {
			return
		}
		fmt.Fprint(buf, "\n\n")
	}

	if !s.opts.NoExports {
		s.RenderExports(buf)
	}

	return
}

func (s *StructToTS) RenderExports(w io.Writer) (err error) {
	io.WriteString(w, "// exports\n")

	export := func(n string) { _, err = fmt.Fprintf(w, "%s%s,\n", s.opts.indents[1], n) }
	if s.opts.ES6 {
		fmt.Fprintf(w, "if (typeof exports === 'undefined') var exports = {};\n\n")
		export = func(n string) { _, err = fmt.Fprintf(w, "exports.%s = %s;\n", n, n) }
	} else {
		io.WriteString(w, "export {\n")
	}

	for _, st := range s.structs {
		export(st.Name)
	}

	if !s.opts.NoHelpers {
		for _, n := range []string{"ParseDate", "ParseNumber", "FromArray", "ToObject"} {
			export(n)
		}
	}

	if !s.opts.ES6 {
		io.WriteString(w, "};\n")
	}
	return
}

func indirect(t reflect.Type) reflect.Type {
	k := t.Kind()
	for k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
	}
	return t
}

func isNumber(k reflect.Kind) bool {
	switch k {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Float32, reflect.Float64:

		return true

	default:
		return false
	}
}

func isStruct(t reflect.Type) bool {
	return indirect(t).Kind() == reflect.Struct
}

func stripType(t reflect.Type) string {
	k := t.Kind()
	switch {
	case isNumber(k):
		return "number"
	case k == reflect.String:
		return "string"
	case k == reflect.Bool:
		return "boolean"
	}

	n := t.String()
	if i := strings.IndexByte(n, '.'); i > -1 {
		n = n[i+1:]
	}
	return n
}
