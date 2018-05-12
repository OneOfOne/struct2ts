package struct2ts

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

var zeroValues = map[string]string{
	"number": "0",
	"string": "''",
	"object": "{}",
	"array":  "[]",
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

	_, err = fmt.Fprintln(w, "}")
	return
}

func (s *Struct) RenderFields(opts *Options, w io.Writer) (err error) {
	for _, f := range s.Fields {
		name := f.Name
		if f.IsOptional && opts.MarkOptional {
			name += "?"
		}
		if opts.InterfaceOnly || opts.NoAssignDefaults {
			_, err = fmt.Fprintf(w, "%s%s: %s;\n", opts.indents[1], name, f.Type(opts.NoDate, false))
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
	fmt.Fprintf(w, "%sif (typeof data !== 'object') return;\n\n", opts.indents[2])
	for _, f := range s.Fields {
		switch t := f.Type(opts.NoDate, true); t {
		case "Date":
			// convert to js date
			fmt.Fprintf(w, "%sif ('%s' in data) this.%s = new Date(data.%s * 1000);\n", opts.indents[2], f.Name, f.Name, f.Name)
		case f.ValType: // struct
			fmt.Fprintf(w, "%sif ('%s' in data) this.%s = new %s(data.%s);\n", opts.indents[2], f.Name, f.Name, f.ValType, f.Name)
		default:
			fmt.Fprintf(w, "%sif ('%s' in data) this.%s = data.%s as %s;\n", opts.indents[2], f.Name, f.Name, f.Name, t)
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
	IsDate     bool
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

func (f *Field) DefaultValue() (def string) {
	if f.CanBeNull {
		return "null"
	}

	if def = zeroValues[f.TsType]; def != "" {
		return def
	}

	if f.TsType == "object" {
		def = "new " + f.ValType + "()"
	}
	return
}

func (f *Field) setProps(sf reflect.StructField) (ignore bool) {
	t := strings.Split(sf.Tag.Get("json"), ",")
	if len(t) == 0 {
		t = strings.Split(sf.Tag.Get("ts"), ",")
	}

	if len(t) == 0 {
		f.Name = sf.Name
		return
	}

	if ignore = t[0] == "-" || t[0] == ""; ignore {
		return true
	}

	f.Name = t[0]
	f.IsDate = sf.Type.Kind() == reflect.Int64 && strings.Contains(strings.ToLower(f.Name), "ts")

	if len(t) == 2 {
		f.IsOptional = strings.Contains(t[1], "omitempty") || strings.Contains(t[1], "optional")
		f.IsDate = f.IsDate || strings.Contains(t[1], "date")
	}

	return
}
