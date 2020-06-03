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
	structs        []*structRep
	seen           map[reflect.Type]*structRep
	anonymousCount int
}

// New returns a new *StructToTS.
func New() *Go2TS {
	ret := &Go2TS{
		seen: map[reflect.Type]*structRep{},
	}
	return ret
}

// Add a struct that needs a TypeScript definition.
func (g *Go2TS) Add(v interface{}) *structRep {
	return g.AddWithName(v, "")
}

// AddWithName adds a struct that needs a TypeScript definition, the TypeScript
// interface name is supplied.
func (g *Go2TS) AddWithName(v interface{}, name string) *structRep {
	var reflectType reflect.Type
	switch v := v.(type) {
	case reflect.Type:
		reflectType = v
	case reflect.Value:
		reflectType = v.Type()
	default:
		reflectType = reflect.TypeOf(v)
	}

	return g.addType(reflectType, name)
}

// Render the TypeScript definitions to the given io.Writer.
func (g *Go2TS) Render(w io.Writer) error {
	for _, st := range g.structs {
		if err := st.render(w); err != nil {
			return err
		}
	}

	return nil
}

// setType fills out the 'partialType' for the given 'typ'.
//
// It will recursively add subTypes to the partialType until the type is
// completely described.
func (g *Go2TS) setType(partialType *partialTypeRep, typ reflect.Type) {
	kind := typ.Kind()
	if kind == reflect.Ptr {
		partialType.canBeNull = true
		typ = removeIndirection(typ)
		kind = typ.Kind()
	}

	// Default to setting the tsType from the Go type.
	partialType.tsType = reflectTypeToTypeScriptType(typ)

	// Update the type if the kind points to something besides a primitive type.
	switch kind {
	case reflect.Map:
		partialType.tsType = "map"
		partialType.keyType = reflectTypeToTypeScriptType(typ.Key())
		partialType.valType = reflectTypeToTypeScriptType(typ.Elem())

		partialType.subType = &partialTypeRep{}
		g.setType(partialType.subType, typ.Elem())

	case reflect.Slice, reflect.Array:
		partialType.canBeNull = kind == reflect.Slice
		partialType.tsType = "array"
		partialType.valType = reflectTypeToTypeScriptType(typ.Elem())

		partialType.subType = &partialTypeRep{}
		g.setType(partialType.subType, typ.Elem())

	case reflect.Struct:
		if partialType.isTime {
			break
		}
		partialType.tsType = "object"
		partialType.valType = g.addType(typ, "").Name

	case reflect.Interface:
		partialType.tsType = "any"
		partialType.valType = ""
	}
}

func (g *Go2TS) addTypeFields(out *structRep, reflectType reflect.Type) {
	for i := 0; i < reflectType.NumField(); i++ {
		structField := reflectType.Field(i)

		// Skip unexported fields.
		if len(structField.Name) == 0 || !ast.IsExported(structField.Name) {
			continue
		}

		var field fieldRep

		structFieldType := structField.Type
		kind := structFieldType.Kind()

		if kind == reflect.Ptr {
			field.canBeNull = true
			// Remove indirection until we get to a non-pointer kind.
			structFieldType = removeIndirection(structFieldType)
			kind = structFieldType.Kind()
		}

		field.initializeFromReflectStructField(structField)

		g.setType(&field.partialTypeRep, structFieldType)

		out.Fields = append(out.Fields, &field)
	}
}

func (g *Go2TS) getAnonymous() string {
	g.anonymousCount++
	return fmt.Sprintf("Anonymous%d", g.anonymousCount)
}

func (g *Go2TS) addType(t reflect.Type, name string) *structRep {
	var ret *structRep

	t = removeIndirection(t)
	if ret = g.seen[t]; ret != nil {
		return ret
	}

	if name == "" {
		name = strings.Title(t.Name())
	}
	if name == "" {
		name = g.getAnonymous()
	}

	ret = &structRep{
		Name:   name,
		Fields: make([]*fieldRep, 0, t.NumField()),
	}

	g.seen[t] = ret
	g.addTypeFields(ret, t)
	g.structs = append(g.structs, ret)
	return ret
}

// partialTypeRep represents either a type of a field, like "string", or part of
// a more complex type like the map[string] part of map[string]time.Time.
type partialTypeRep struct {
	// tsType is the TypeScript type.
	tsType string

	// keyType is the type of the key if this tsType is a map.
	keyType string

	// valType is the type of the value if this tsType is a "map", the type of
	// the thing in the array if tsType is "array", and the name of the
	// TypeScript interface if the tsType is "object".
	valType string

	// isTime is true if the struct is actually a time.Time, which we handle
	// specially as a string.
	isTime    bool
	canBeNull bool
	subType   *partialTypeRep
}

// typeScriptSubType returns the partialType formatted as TypeScript.
func (s *partialTypeRep) typeScriptSubType() string {
	var ret string
	switch s.tsType {
	case "array":
		ret = s.subType.typeScriptSubType() + "[]"
	case "map":
		ret = fmt.Sprintf("{ [key: %s]: %s }", s.keyType, s.subType.typeScriptSubType())
	case "object":
		ret = s.valType
	default:
		ret = s.tsType
	}

	return ret
}

// fieldRep represents one field in a struct.
type fieldRep struct {
	partialTypeRep
	name       string
	isOptional bool
}

func (f *fieldRep) typeScriptType() string {
	ret := f.typeScriptSubType()

	if f.canBeNull {
		ret += " | null"
	}

	return ret
}

func (f *fieldRep) String() string {
	optional := ""
	if f.isOptional {
		optional = "?"
	}

	return fmt.Sprintf("\t%s%s: %s;", f.name, optional, f.typeScriptType())
}

func (f *fieldRep) initializeFromReflectStructField(structField reflect.StructField) {
	reflectType := structField.Type
	jsonTag := strings.Split(structField.Tag.Get("json"), ",")

	f.name = structField.Name
	if len(jsonTag) > 0 && jsonTag[0] != "" {
		f.name = jsonTag[0]
	}

	f.isTime = isTime(reflectType)
	f.isOptional = f.isOptional || len(jsonTag) > 1 && jsonTag[1] == "omitempty"
	f.tsType = reflectTypeToTypeScriptType(reflectType)
}

func isTime(t reflect.Type) bool {
	return t.Name() == "Time" && t.PkgPath() == "time"
}

// structRep represents a single Go struct.
type structRep struct {
	Name   string
	Fields []*fieldRep
}

var structRepTemplate = template.Must(template.New("").Parse(`
export interface {{ .Name }} {
{{ range .Fields -}}
	{{- . }}
{{ end -}}
}
`))

func (s *structRep) render(w io.Writer) (err error) {
	return structRepTemplate.Execute(w, s)
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

	// Falling through to use the native type, which works for []string, and
	// also helps figure out the error.
	nativeType := typ.String()
	if i := strings.IndexByte(nativeType, '.'); i > -1 {
		nativeType = nativeType[i+1:]
	}
	if strings.HasPrefix(nativeType, "[]") {
		return nativeType[2:]
	}
	return nativeType
}
