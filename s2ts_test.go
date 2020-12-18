package struct2ts_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/OneOfOne/struct2ts"
)

type OtherStruct struct {
	T time.Time `json:"t,omitempty"`
}

type ComplexStruct struct {
	S           string       `json:"s,omitempty"`
	I           int          `json:"i,omitempty"`
	F           float64      `json:"f,omitempty"`
	TS          *int64       `json:"ts,omitempty" ts:"date,null"`
	T           time.Time    `json:"t,omitempty"` // automatically handled
	NullOther   *OtherStruct `json:"o,omitempty"`
	NoNullOther *OtherStruct `json:"nno,omitempty" ts:",no-null"`
	Data        Data         `json:"d"`
	DataPtr     *Data        `json:"dp"`
	CT          CustomType   `json:"ct"`
}

type Data map[string]interface{}

func TestComplexStruct(t *testing.T) {
	s2ts := struct2ts.New(nil)
	s2ts.Add(ComplexStruct{})

	var buf bytes.Buffer

	s2ts.RenderTo(&buf)
	// Output:
	// // helpers
	// const maxUnixTSInSeconds = 9999999999;
	//
	// function ParseDate(d: Date | number | string): Date {
	// 	if (d instanceof Date) return d;
	// 	if (typeof d === 'number') {
	// 		if (d > maxUnixTSInSeconds) return new Date(d);
	// 		return new Date(d * 1000); // go ts
	// 	}
	// 	return new Date(d);
	// }
	//
	// function ParseNumber(v: number | string, isInt = false): number {
	// 	if (!v) return 0;
	// 	if (typeof v === 'number') return v;
	// 	return (isInt ? parseInt(v) : parseFloat(v)) || 0;
	// }
	//
	// function FromArray<T>(Ctor: { new(v: any): T }, data?: any[] | any, def = null): T[] | null {
	// 	if (!data || !Object.keys(data).length) return def;
	// 	const d = Array.isArray(data) ? data : [data];
	// 	return d.map((v: any) => new Ctor(v));
	// }
	//
	// function ToObject(o: any, typeOrCfg: any = {}, child = false): any {
	// 	if (!o) return null;
	// 	if (typeof o.toObject === 'function' && child) return o.toObject();
	//
	// 	switch (typeof o) {
	// 		case 'string':
	// 			return typeOrCfg === 'number' ? ParseNumber(o) : o;
	// 		case 'boolean':
	// 		case 'number':
	// 			return o;
	// 	}
	//
	// 	if (o instanceof Date) {
	// 		return typeOrCfg === 'string' ? o.toISOString() : Math.floor(o.getTime() / 1000);
	// 	}
	//
	// 	if (Array.isArray(o)) return o.map((v: any) => ToObject(v, typeOrCfg, true));
	//
	// 	const d: any = {};
	//
	// 	for (const k of Object.keys(o)) {
	// 		const v: any = o[k];
	// 		if (!v) continue;
	// 		d[k] = ToObject(v, typeOrCfg[k] || {}, true);
	// 	}
	//
	// 	return d;
	// }
	//
	// // classes
	// // struct2ts:github.com/OneOfOne/struct2ts_test.ComplexStructOtherStruct
	// class ComplexStructOtherStruct {
	// 	t: Date;
	//
	// 	constructor(data?: any) {
	// 		const d: any = (data && typeof data === 'object') ? ToObject(data) : {};
	// 		this.t = ('t' in d) ? ParseDate(d.t) : new Date();
	// 	}
	//
	// 	toObject(): any {
	// 		const cfg: any = {};
	// 		cfg.t = 'string';
	// 		return ToObject(this, cfg);
	// 	}
	// }
	//
	// // struct2ts:github.com/OneOfOne/struct2ts_test.ComplexStruct
	// class ComplexStruct {
	// 	s: string;
	// 	i: number;
	// 	f: number;
	// 	ts: Date | null;
	// 	t: Date;
	// 	o: ComplexStructOtherStruct | null;
	// 	nno: ComplexStructOtherStruct;
	// 	d: { [key: string]: any };
	// 	dp: { [key: string]: any } | null;
	//
	// 	constructor(data?: any) {
	// 		const d: any = (data && typeof data === 'object') ? ToObject(data) : {};
	// 		this.s = ('s' in d) ? d.s as string : '';
	// 		this.i = ('i' in d) ? d.i as number : 0;
	// 		this.f = ('f' in d) ? d.f as number : 0;
	// 		this.ts = ('ts' in d) ? ParseDate(d.ts) : null;
	// 		this.t = ('t' in d) ? ParseDate(d.t) : new Date();
	// 		this.o = ('o' in d) ? new ComplexStructOtherStruct(d.o) : null;
	// 		this.nno = new ComplexStructOtherStruct(d.nno);
	// 		this.d = ('d' in d) ? d.d as { [key: string]: any } : {};
	// 		this.dp = ('dp' in d) ? d.dp as { [key: string]: any } : null;
	// 	}
	//
	// 	toObject(): any {
	// 		const cfg: any = {};
	// 		cfg.i = 'number';
	// 		cfg.f = 'number';
	// 		cfg.t = 'string';
	// 		return ToObject(this, cfg);
	// 	}
	// }
	//
	// // exports
	// export {
	// 	ComplexStructOtherStruct,
	// 	ComplexStruct,
	// 	ParseDate,
	// 	ParseNumber,
	// 	FromArray,
	// 	ToObject,
	// };
}

type CustomType struct {
	Value string
}

func (c *CustomType) CustomTypeScriptType() string {
	return "string"
}

func (s *CustomType) UnmarshalText(text []byte) error {
	s.Value = string(text)
	return nil
}

func (s CustomType) MarshalText() ([]byte, error) {
	return []byte(s.Value), nil
}
