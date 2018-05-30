# struct2ts [![GoDoc](https://godoc.org/github.com/OneOfOne/struct2ts?status.svg)](https://godoc.org/github.com/OneOfOne/struct2ts)

An extremely simple and powerful Go struct to Typescript Class generator.

Inspired by [tkrajina/typescriptify-golang-structs](https://github.com/tkrajina/typescriptify-golang-structs).

## Install

	go get -u -v github.com/OneOfOne/struct2ts/...

## Features

* Fairly decent command line interface if you don't wanna write a generator yourself.
* Automatically handles Go `int64` timestamps `<->` Javascript `Date`.
* Automatically handles json tags.

## Options

### There's an extra struct tag to control the output, `ts`, valid options are

* `-` omit this field.
* `date` handle converting `time.Time{}.Unix() <-> javascript Date`.
* `,no-null` only valid for struct fields, forces creating a new class rather than using `null` in TS.
* `,null` allows any field type to be `null`.

## Example

* Input:

```go
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
}
```

* Output:

```ts
// ... helpers...
// struct2ts:github.com/OneOfOne/struct2ts_test.ComplexStructOtherStruct
class ComplexStructOtherStruct {
	t: Date;

	constructor(data?: any) {
		const d: any = (data && typeof data === 'object') ? ToObject(data) : {};
		this.t = ('t' in d) ? ParseDate(d.t) : new Date();
	}

	toObject(): any {
		const cfg: any = {};
		cfg.t = 'string';
		return ToObject(this, cfg);
	}
}

// struct2ts:github.com/OneOfOne/struct2ts_test.ComplexStruct
class ComplexStruct {
	s: string;
	i: number;
	f: number;
	ts: Date | null;
	t: Date;
	o: ComplexStructOtherStruct | null;
	nno: ComplexStructOtherStruct;

	constructor(data?: any) {
		const d: any = (data && typeof data === 'object') ? ToObject(data) : {};
		this.s = ('s' in d) ? d.s as string : '';
		this.i = ('i' in d) ? d.i as number : 0;
		this.f = ('f' in d) ? d.f as number : 0;
		this.ts = ('ts' in d) ? ParseDate(d.ts) : null;
		this.t = ('t' in d) ? ParseDate(d.t) : new Date();
		this.o = ('o' in d) ? new ComplexStructOtherStruct(d.o) : null;
		this.nno = new ComplexStructOtherStruct(d.nno);
	}

	toObject(): any {
		const cfg: any = {};
		cfg.i = 'number';
		cfg.f = 'number';
		cfg.t = 'string';
		return ToObject(this, cfg);
	}
}
// ...exports...

```

## Command Line Usage

```
➤ struct2ts -h
usage: struct2ts [<flags>] [<pkg.struct>...]

Flags:
	-h, --help                  Show context-sensitive help (also try --help-long
								and --help-man).
		--indent="\t"           Output indentation.
	-m, --mark-optional-fields  Add `?` to fields with omitempty.
	-6, --es6                   generate es6 code
	-C, --no-ctor               Don't generate a ctor.
	-T, --no-toObject           Don't generate a Class.toObject() method.
	-E, --no-exports            Don't automatically export the generated types.
	-D, --no-date               Don't automatically handle time.Unix () <-> JS
								Date().
	-H, --no-helpers            Don't output the helpers.
	-N, --no-default-values     Don't assign default/zero values in the ctor.
	-i, --interface             Only generate an interface (disables all the other
								options).
	-s, --src-only              Only output the Go code (helpful if you want to
								edit it yourself).
	-p, --package-name="main"   the package name to use if --src-only is set.
	-k, --keep-temp             Keep the generated Go file, ignored if --src-only
								is set.
	-o, --out="-"               Write the output to a file instead of stdout.
	-V, --version               Show application version.

Args:
	[<pkg.struct>]  List of structs to convert (github.com/you/auth/users.User,
					users.User or users.User:AliasUser).

```

## TODO

* Use [xast](https://github.com/OneOfOne/xast) to skip reflection.
* Support annoymous structs.
* ~~Support ES6.~~

## License

This project is released under the [BSD 3-clause "New" or "Revised" License](https://github.com/golang/go/blob/master/LICENSE).
