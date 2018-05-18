package struct2ts

import (
	"fmt"
	"go/ast"
	"io"
	"log"
	"reflect"
	"strings"
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
	Fields []Field

	t reflect.Type
}

func (s *Struct) RenderTo(opts *Options, w io.Writer) (err error) {
	if _, err = fmt.Fprintf(w, "// struct2ts:%s.%s\n", s.t.PkgPath(), s.Name); err != nil {
		return
	}

	if opts.InterfaceOnly {
		if _, err = fmt.Fprintf(w, "export interface %s {\n", s.Name); err != nil {
			return
		}
	} else {
		if _, err = fmt.Fprintf(w, "export class %s {\n", s.Name); err != nil {
			return
		}
	}

	if err = s.RenderFields(opts, w); err != nil {
		return
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
		name := f.Name
		if f.IsOptional && opts.MarkOptional {
			name += "?"
		}
		if opts.InterfaceOnly || opts.NoAssignDefaults || !opts.NoConstructor {
			_, err = fmt.Fprintf(w, "%s%s: %s;\n", opts.indents[1], name, f.Type(opts.NoDate, false))
			continue
		} else {
			_, err = fmt.Fprintf(w, "%s%s: %s = %s;\n", opts.indents[1], name, f.Type(opts.NoDate, false), f.DefaultValue())
		}
	}
	return
}

func (s *Struct) RenderConstructor(opts *Options, w io.Writer) (err error) {
	if opts.NoConstructor || opts.InterfaceOnly {
		return
	}

	fmt.Fprintf(w, "\n%sconstructor(data?: any) {\n", opts.indents[1])
	fmt.Fprintf(w, "%sconst d: any = typeof data === 'object' ? data : {};\n\n", opts.indents[2])
	for _, f := range s.Fields {
		t := f.Type(opts.NoDate, true)
		switch {
		case t == "Date":
			// convert to js date
			fmt.Fprintf(w, "%sthis.%s = ('%s' in d) ? new Date(d.%s * 1000) : %s;\n",
				opts.indents[2], f.Name, f.Name, f.Name, f.DefaultValue())
		case t == f.ValType: // struct
			fmt.Fprintf(w, "%sthis.%s = ('%s' in d) ? new %s(d.%s) : %s;\n",
				opts.indents[2], f.Name, f.Name, f.ValType, f.Name, f.DefaultValue())
		case f.TsType == "array" && !f.IsNative():
			fmt.Fprintf(w, "%sthis.%s = (d.%s || []).map((v: any) => new %s(v));\n",
				opts.indents[2], f.Name, f.Name, f.ValType)
		case f.TsType == "map" && !f.IsNative():
			// fmt.Fprintf(w, "%sthis.%s = Object.keys(d.%s || {}).mp((k: any) => new %s(v));\n",
			// 	opts.indents[2], f.Name, f.Name, f.ValType)
			fallthrough
		default:
			fmt.Fprintf(w, "%sthis.%s = ('%s' in d) ? d.%s as %s : %s;\n",
				opts.indents[2], f.Name, f.Name, f.Name, t, f.DefaultValue())
		}
	}
	_, err = fmt.Fprintf(w, "%s}\n", opts.indents[1])

	return
}

func (s *Struct) RenderToObject(opts *Options, w io.Writer) (err error) {
	if opts.NoToObject || opts.InterfaceOnly {
		return
	}
	fmt.Fprintf(w, "\n%stoObject(): { [key:string]: any } {\n", opts.indents[1])
	fmt.Fprintf(w, "%sconst data: { [key:string]: any } = {};\n\n", opts.indents[2])
	for _, f := range s.Fields {
		switch t := f.Type(opts.NoDate, true); t {
		case "Date":
			// convert to valid go time
			fmt.Fprintf(w, "%sif (this.%s) data.%s = this.%s.getTime() / 1000 >>> 0;\n", opts.indents[2], f.Name, f.Name, f.Name)
		case f.ValType: // struct
			fmt.Fprintf(w, "%sif (this.%s) data.%s = this.%s.toObject();\n", opts.indents[2], f.Name, f.Name, f.Name)
		default:
			fmt.Fprintf(w, "%sif (this.%s) data.%s = this.%s;\n", opts.indents[2], f.Name, f.Name, f.Name)
		}
	}
	_, err = fmt.Fprintf(w, "%sreturn data;\n%s}\n", opts.indents[2], opts.indents[1])
	return
}

type Field struct {
	Name       string `json:"name"`
	TsType     string `json:"type"`
	KeyType    string `json:"keyType,omitempty"`
	ValType    string `json:"valType,omitempty"`
	CanBeNull  bool   `json:"canBeNull"`
	IsOptional bool   `json:"isOptional"`
	IsDate     bool   `json:"isDate"`
}

func (f *Field) Type(noDate, noSuffix bool) (out string) {
	switch out = f.TsType; out {
	case "string", "number", "boolean":
	case "array":
		out = f.ValType + "[]"
	case "map":
		out = fmt.Sprintf("{ [key: %s]: %s }", f.KeyType, f.ValType)
	case "object":
		if out = f.ValType; out == "" {
			out = "any"
		}
	}

	if f.IsDate && !noDate {
		out = "Date"
	}

	if !noSuffix && f.CanBeNull {
		out += " | null"
	}

	return
}

func (f *Field) DefaultValue() string {
	if f.CanBeNull {
		return "null"
	}

	if f.IsDate {
		return "new Date()"
	}

	if f.TsType == "object" && f.ValType != "" {
		return "new " + f.ValType + "()"
	}

	return zeroValues[f.TsType]
}

func (f *Field) IsNative() bool {
	switch f.TsType {
	case "array", "map":
		return IsNative(f.ValType)
	default:
		return IsNative(f.TsType)
	}
}
func (f *Field) setProps(sf reflect.StructField, sft reflect.Type) (ignore bool) {
	if len(sf.Name) > 0 && !ast.IsExported(sf.Name) {
		return true
	}

	if sf.Anonymous {
		log.Println("anonymous structs aren't supported, yet.")
		return true
	}

	var (
		jsonTag = strings.Split(sf.Tag.Get("json"), ",")
		tsTag   = strings.Split(sf.Tag.Get("ts"), ",")
	)

	if ignore = len(tsTag) > 0 && tsTag[0] == "-" || len(jsonTag) > 0 && jsonTag[0] == "-"; ignore {
		return
	}

	if len(jsonTag) == 0 || len(jsonTag) == 1 && jsonTag[0] == "" {
		f.Name = sf.Name
		return
	}

	f.Name = jsonTag[0]
	f.IsDate = len(tsTag) > 0 && tsTag[0] == "date" || sft.Kind() == reflect.Int64 && strings.HasSuffix(f.Name, "TS")

	if len(tsTag) > 1 {
		switch tsTag[1] {
		case "no-null":
			f.CanBeNull = false
		case "null":
			f.CanBeNull = true
		case "optional":
			f.IsOptional = true
		}
	}

	f.IsOptional = f.IsOptional || len(jsonTag) > 1 && jsonTag[1] == "omitempty"
	f.TsType = stripType(sft)

	return
}

func IsNative(t string) bool {
	switch t {
	case "string", "number", "boolean", "Date":
		return true
	default:
		return false
	}
}
