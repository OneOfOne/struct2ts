package go2ts

import (
	"bytes"
	"image/color"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SomeStruct struct {
	B string
}

type OtherStruct struct {
	T time.Time `json:"t,omitempty"`
}

type Mode string

type Data map[string]interface{}

type ComplexStruct struct {
	String                string `json:"s"`
	Bool                  bool   `json:"b"`
	NoTypeAnnotation      string
	Int                   int                          `json:"i"`
	Float64               float64                      `json:"f"`
	Time                  time.Time                    `json:"t"`
	Other                 *OtherStruct                 `json:"o"`
	OptionalString        string                       `json:"se,omitempty"`
	OptionalInt           int                          `json:"ie,omitempty"`
	OptionalFloat64       float64                      `json:"fe,omitempty"`
	OptionalTime          time.Time                    `json:"te,omitempty"`
	OptionalOther         *OtherStruct                 `json:"oe,omitempty"`
	Data                  Data                         `json:"d"`
	DataPtr               *Data                        `json:"dp"`
	MapStringSlice        map[string][]*string         `json:"mss"`
	MapIntKeys            map[int]string               `json:"mik"`
	Slice                 []string                     `json:"slice"`
	SliceOfSlice          [][]string                   `json:"sos"`
	SliceOfData           []Data                       `json:"sod"`
	MapOfData             map[string]Data              `json:"mod"`
	MapOfSliceOfData      map[string][]Data            `json:"mosod"`
	MapOfMapOfSliceOfData map[string]map[string][]Data `json:"momosod"`
	Mode                  Mode                         `json:"mode"`
	InlineStruct          struct{ A int }              `json:"inline"`
	Array                 [3]string                    `json:"array"`
	skipped               bool
	Color                 color.Alpha
}

const complexStructExpected = `
export interface OtherStruct {
	t?: string;
}

export interface Anonymous1 {
	A: number;
}

export interface Alpha {
	A: number;
}

export interface ComplexStruct {
	s: string;
	b: boolean;
	NoTypeAnnotation: string;
	i: number;
	f: number;
	t: string;
	o: OtherStruct | null;
	se?: string;
	ie?: number;
	fe?: number;
	te?: string;
	oe?: OtherStruct | null;
	d: { [key: string]: any };
	dp: { [key: string]: any } | null;
	mss: { [key: string]: string[] };
	mik: { [key: number]: string };
	slice: string[] | null;
	sos: string[][] | null;
	sod: { [key: string]: any }[] | null;
	mod: { [key: string]: { [key: string]: any } };
	mosod: { [key: string]: { [key: string]: any }[] };
	momosod: { [key: string]: { [key: string]: { [key: string]: any }[] } };
	mode: string;
	inline: Anonymous1;
	array: string[];
	Color: Alpha;
}
`

func TestRender_ComplexStruct_Success(t *testing.T) {
	go2ts := New()
	err := go2ts.Add(ComplexStruct{})
	require.NoError(t, err)
	var b bytes.Buffer
	go2ts.Render(&b)
	assert.Equal(t, complexStructExpected, b.String())
}

func TestRender_NoTypesAdded_ReturnsEmptyString(t *testing.T) {
	go2ts := New()
	var b bytes.Buffer
	go2ts.Render(&b)
	assert.Equal(t, "", b.String())
}

func TestRender_SameTypeAddedInMultipleWays_RendersTypeOnce(t *testing.T) {
	go2ts := New()
	err := go2ts.Add(reflect.TypeOf(SomeStruct{}))
	require.NoError(t, err)
	err = go2ts.AddWithName(SomeStruct{}, "ADifferentName")
	require.NoError(t, err)
	err = go2ts.Add(reflect.TypeOf([]SomeStruct{}).Elem())
	require.NoError(t, err)
	err = go2ts.Add(SomeStruct{})
	require.NoError(t, err)
	err = go2ts.Add(&SomeStruct{})
	require.NoError(t, err)
	err = go2ts.Add(reflect.New(reflect.TypeOf(SomeStruct{})))
	require.NoError(t, err)
	var b bytes.Buffer
	go2ts.Render(&b)
	expected := `
export interface SomeStruct {
	B: string;
}
`
	assert.Equal(t, expected, b.String())
}

func TestRender_FirstAddDeterminesInterfaceName(t *testing.T) {
	go2ts := New()
	err := go2ts.AddWithName(SomeStruct{}, "ADifferentName")
	require.NoError(t, err)
	err = go2ts.Add(reflect.TypeOf(SomeStruct{}))
	require.NoError(t, err)
	var b bytes.Buffer
	go2ts.Render(&b)
	expected := `
export interface ADifferentName {
	B: string;
}
`
	assert.Equal(t, expected, b.String())
}

func TestAdd_NonStructType_ReturnsError(t *testing.T) {
	go2ts := New()
	err := go2ts.Add("this is not a struct")
	assert.Error(t, err)
	assert.Equal(t, "string is not a struct.", err.Error())
}
