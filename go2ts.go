// Package go2ts is a module for generating TypeScript definitions from Go
// structs.
package go2ts

import (
	"fmt"
	"go/ast"
	"html/template"
	"io"
	"reflect"
	"strings"
)

// Go2TS writes TypeScript interface definitions for the given Go structs.
type Go2TS struct {
	structs        []structRep
	seen           map[reflect.Type]structRep
	anonymousCount int
}

// New returns a new *StructToTS.
func New() *Go2TS {
	ret := &Go2TS{
		seen: map[reflect.Type]structRep{},
	}
	return ret
}

// Add a struct that needs a TypeScript definition.
//
// Just a wrapper for AddWithName with an interfaceName of "".
//
// See AddWithName.
func (g *Go2TS) Add(v interface{}) error {
	return g.AddWithName(v, "")
}

// AddWithName adds a struct that needs a TypeScript definition.
//
// The value passed in must resolve to a struct, a reflect.Type, or a
// reflect.Value of a struct. That is, a string or number for v will cause
// AddWithName to return an error, but a pointer to a struct is fine.
//
// The 'name' supplied will be the TypeScript interface name.  If
// 'interfaceName' is "" then the struct name will be used. If the struct is
// anonymous it will be given a name of the form "AnonymousN".
//
// The fields of the struct will be named following the convention for json
// serialization, including using the json tag if supplied.
//
// Fields tagged with `json:",omitempty"` will have "| null" added to their
// type.
//
// There is special handling of time.Time types to be TypeScript "string"s since
// they implement MarshalJSON, see
// https://pkg.go.dev/time?tab=doc#Time.MarshalJSON.
func (g *Go2TS) AddWithName(v interface{}, interfaceName string) error {
	var reflectType reflect.Type
	switch v := v.(type) {
	case reflect.Type:
		reflectType = v
	case reflect.Value:
		reflectType = v.Type()
	default:
		reflectType = reflect.TypeOf(v)
	}

	_, err := g.addType(reflectType, interfaceName)
	return err
}

// Render the TypeScript definitions to the given io.Writer.
func (g *Go2TS) Render(w io.Writer) {
	for _, st := range g.structs {
		st.render(w)
	}
}

// populateTypeDetails fills out the 'partialType' for the given 'typ'.
//
// It will recursively add subTypes to the partialType until the type is
// completely described.
func (g *Go2TS) tSTypeFromStructFieldType(reflectType reflect.Type) *tsType {
	var ret tsType
	kind := reflectType.Kind()
	if kind == reflect.Ptr {
		ret.canBeNull = true
		reflectType = removeIndirection(reflectType)
		kind = reflectType.Kind()
	}

	// Default to setting the tsType from the Go type.
	ret.typeName = reflectTypeToTypeScriptType(reflectType)

	// Update the type if the kind points to something besides a primitive type.
	switch kind {
	case reflect.Map:
		ret.typeName = "map"
		ret.keyType = reflectTypeToTypeScriptType(reflectType.Key())

		ret.subType = g.tSTypeFromStructFieldType(reflectType.Elem())

	case reflect.Slice, reflect.Array:
		ret.typeName = "array"
		ret.canBeNull = (kind == reflect.Slice)

		ret.subType = g.tSTypeFromStructFieldType(reflectType.Elem())

	case reflect.Struct:
		if isTime(reflectType) {
			break
		}
		ret.typeName = "interface"
		st, _ := g.addType(reflectType, "")
		ret.interfaceName = st.Name

	case reflect.Interface:
		ret.typeName = "any"
	}
	return &ret
}

func (g *Go2TS) addTypeFields(st *structRep, reflectType reflect.Type) {
	for i := 0; i < reflectType.NumField(); i++ {
		structField := reflectType.Field(i)

		// Skip unexported fields.
		if len(structField.Name) == 0 || !ast.IsExported(structField.Name) {
			continue
		}

		structFieldType := structField.Type
		field := newFieldRep(structField)
		field.tsType = g.tSTypeFromStructFieldType(structFieldType)
		st.Fields = append(st.Fields, field)
	}
}

func (g *Go2TS) getAnonymousInterfaceName() string {
	g.anonymousCount++
	return fmt.Sprintf("Anonymous%d", g.anonymousCount)
}

func (g *Go2TS) addType(reflectType reflect.Type, interfaceName string) (structRep, error) {
	var ret structRep

	reflectType = removeIndirection(reflectType)
	var ok bool
	if ret, ok = g.seen[reflectType]; ok {
		return ret, nil
	}

	if reflectType.Kind() != reflect.Struct {
		return structRep{}, fmt.Errorf("%s is not a struct.", reflectType)
	}

	if interfaceName == "" {
		interfaceName = strings.Title(reflectType.Name())
	}
	if interfaceName == "" {
		interfaceName = g.getAnonymousInterfaceName()
	}

	ret = structRep{
		Name:   interfaceName,
		Fields: make([]fieldRep, 0, reflectType.NumField()),
	}

	g.seen[reflectType] = ret
	g.addTypeFields(&ret, reflectType)
	g.structs = append(g.structs, ret)
	return ret, nil
}

// tsType represents either a type of a field, like "string", or part of
// a more complex type like the map[string] part of map[string]time.Time.
type tsType struct {
	// typeName is the TypeScript type, such as "string", or "map", or "SomeInterfaceName".
	typeName string

	// keyType is the type of the key if this tsType is a map, such as "string" or "number".
	keyType string

	// interfaceName is the name of the TypeScript interface if the tsType is
	// "interface".
	interfaceName string

	canBeNull bool
	subType   *tsType
}

// String returns the tsType formatted as TypeScript.
func (s tsType) String() string {
	var ret string
	switch s.typeName {
	case "array":
		ret = s.subType.String() + "[]"
	case "map":
		ret = fmt.Sprintf("{ [key: %s]: %s }", s.keyType, s.subType.String())
	case "interface":
		ret = s.interfaceName
	default:
		ret = s.typeName
	}

	return ret
}

// fieldRep represents one field in a struct.
type fieldRep struct {
	// The name of the interface field.
	name       string
	tsType     *tsType
	isOptional bool
}

func (f fieldRep) String() string {
	optional := ""
	if f.isOptional {
		optional = "?"
	}

	canBeNull := ""
	if f.tsType.canBeNull {
		canBeNull = " | null"
	}

	return fmt.Sprintf("\t%s%s: %s%s;", f.name, optional, f.tsType.String(), canBeNull)
}

// newFieldRep creates a new fieldRep from the given reflect.StructField.
func newFieldRep(structField reflect.StructField) fieldRep {
	var ret fieldRep
	jsonTag := strings.Split(structField.Tag.Get("json"), ",")

	ret.name = structField.Name
	if len(jsonTag) > 0 && jsonTag[0] != "" {
		ret.name = jsonTag[0]
	}
	ret.isOptional = len(jsonTag) > 1 && jsonTag[1] == "omitempty"
	return ret
}

func isTime(t reflect.Type) bool {
	return t.Name() == "Time" && t.PkgPath() == "time"
}

// structRep represents a single Go struct.
type structRep struct {
	// Name is the TypeScript interface Name in the generated output.
	Name   string
	Fields []fieldRep
}

var structRepTemplate = template.Must(template.New("").Parse(`
export interface {{ .Name }} {
{{ range .Fields -}}
	{{- . }}
{{ end -}}
}
`))

func (s *structRep) render(w io.Writer) {
	if err := structRepTemplate.Execute(w, s); err != nil {
		panic("Template is invalid.")
	}
}

func removeIndirection(reflectType reflect.Type) reflect.Type {
	kind := reflectType.Kind()
	// Follow all the pointers until we get to a non-Ptr kind.
	for kind == reflect.Ptr {
		reflectType = reflectType.Elem()
		kind = reflectType.Kind()
	}
	return reflectType
}

func reflectTypeToTypeScriptType(typ reflect.Type) string {
	if isTime(typ) {
		return "string"
	}

	kind := typ.Kind()
	switch kind {
	case reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uint,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Int,
		reflect.Float32,
		reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	}

	// Falling through to use the native type.
	nativeType := typ.String()
	// Drop the package prefix.
	if i := strings.IndexByte(nativeType, '.'); i > -1 {
		nativeType = nativeType[i+1:]
	}
	return nativeType
}
