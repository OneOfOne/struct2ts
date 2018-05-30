package struct2ts

import (
	"fmt"
	"go/ast"
	"io"
	"log"
	"reflect"
	"strings"
)

type Field struct {
	Name       string `json:"name"`
	TsType     string `json:"type"`
	KeyType    string `json:"keyType,omitempty"`
	ValType    string `json:"valType,omitempty"`
	CanBeNull  bool   `json:"canBeNull"`
	IsOptional bool   `json:"isOptional"`
	IsDate     bool   `json:"isDate"`
}

func (f *Field) Type(opts *Options, noSuffix bool) (out string) {
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

	if f.IsDate && !opts.NoDate {
		out = "Date"
	}

	if !noSuffix && f.CanBeNull {
		out += " | null"
	}

	return
}

func (f *Field) RenderTopLevel(w io.Writer, opts *Options) (err error) {
	name, t := f.Name, f.Type(opts, false)
	if f.IsOptional && opts.MarkOptional {
		name += "?"
	}

	io.WriteString(w, opts.indents[1])
	io.WriteString(w, name+": ")

	if opts.InterfaceOnly || opts.NoAssignDefaults || !opts.NoConstructor {
		_, err = io.WriteString(w, t)
	} else {
		_, err = fmt.Fprintf(w, "%s = %s", t, f.DefaultValue())
	}

	io.WriteString(w, ";\n")

	return
}

func (f *Field) RenderCtor(w io.Writer, opts *Options) (err error) {
	var (
		t            = f.Type(opts, true)
		d            = f.DefaultValue()
		printDefault = true
	)

	io.WriteString(w, opts.indents[2])
	io.WriteString(w, "this.")
	io.WriteString(w, f.Name)
	io.WriteString(w, " = ")

	switch {
	case t == "Date":
		// convert to js date
		_, err = fmt.Fprintf(w, "('%s' in d) ? ParseDate(d.%s)", f.Name, f.Name)
	case t == f.ValType: // struct
		if printDefault = d == "null"; printDefault {
			_, err = fmt.Fprintf(w, "('%s' in d) ? new %s(d.%s)", f.Name, f.ValType, f.Name)
		} else {
			_, err = fmt.Fprintf(w, "new %s(d.%s)", f.ValType, f.Name)
		}
	case f.TsType == "array" && !f.IsNative():
		_, err = fmt.Fprintf(w, "Array.isArray(d.%s) ? d.%s.map((v%s) => new %s(v))",
			f.Name, f.Name, TypeSuffix("any", opts.ES6, false), f.ValType)
	case f.TsType == "map" && !f.IsNative():
		// fmt.Fprintf(w, "Object.keys(d.%s || {}).mp((k: any) => new %s(v));\n",
		//  f.Name, f.ValType)
		fallthrough
	default:
		_, err = fmt.Fprintf(w, "('%s' in d) ? d.%s%s", f.Name, f.Name, TypeSuffix(t, opts.ES6, true))
	}

	if printDefault {
		io.WriteString(w, " : ")
		io.WriteString(w, f.DefaultValue())
	}
	io.WriteString(w, ";\n")

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

	if f.Name = sf.Name; len(jsonTag) > 0 && jsonTag[0] != "" {
		f.Name = jsonTag[0]

	}

	f.IsDate = isDate(sft) || len(tsTag) > 0 && tsTag[0] == "date" || sft.Kind() == reflect.Int64 && strings.HasSuffix(f.Name, "TS")

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

func TypeSuffix(t string, es6, as bool) string {
	if es6 {
		return ""
	}

	if as {
		return " as " + t
	}

	return ": " + t
}

func isDate(t reflect.Type) bool {
	return t.Name() == "Time" && t.PkgPath() == "time"
}
