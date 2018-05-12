# struct2ts [![GoDoc](https://godoc.org/github.com/OneOfOne/struct2ts?status.svg)](https://godoc.org/github.com/OneOfOne/struct2ts)

A extremely simple and powerful Go struct to Typescript Class generator.

Inspired by [tkrajina/typescriptify-golang-structs](https://github.com/tkrajina/typescriptify-golang-structs).

## Install

	go get -u -v github.com/OneOfOne/struct2ts/...

## Features

* Fairly decent command line interface if you don't wanna write a generator yourself.
* Automatically handles Go `int64` timestamps `<->` Javascript `Date`.
* Automatically handles json tags.

## Example

```
┏━ oneofone@Ava ❨✪/O/struct2ts❩ ❨master ⚡❩
┗━━➤ struct2ts -h
usage: struct2ts [<flags>] <pkg.struct>...

Flags:
	-h, --help                  Show context-sensitive help (also try --help-long and --help-man).
		--indent="\t"           Output indentation.
	-m, --mark-optional-fields  Add `?` to fields with omitempty.
	-C, --no-ctor               Don't generate a ctor.
	-T, --no-toObject           Don't generate a Class.toObject() method.
	-D, --no-date               Don't automatically handle time.Unix () <-> JS Date().
	-N, --no-default-values     Don't assign default/zero values in the ctor.
	-i, --interface             Only generate an interface (disables all the other options).
	-k, --keep-temp             Keep the generated temporary Go file.
	-V, --version               Show application version.

Args:
	<pkg.struct>  List of structs to convert (github.com/you/auth/users.User or just users.User).

┏━ oneofone@Ava ❨✪/O/struct2ts❩ ❨master ⚡❩
┗━━➤ struct2ts github.com/OneOfOne/struct2ts.Options # or just struct2ts.Options and /x/imports will handle it.

main.go:75: executing: go run /tmp/s2ts_gen_725543931.go
// struct2ts:github.com/OneOfOne/struct2ts.Options
export class Options {
	Indent: string = '';
	NoAssignDefaults: boolean = false;
	InterfaceOnly: boolean = false;
	MarkOptional: boolean = false;
	NoConstructor: boolean = false;
	NoToObject: boolean = false;
	NoDate: boolean = false;
	indents: string[] = [];

	constructor(data?: any) {
		if (typeof data !== 'object') return;

		if ('Indent' in data) this.Indent = data.Indent as string;
		if ('NoAssignDefaults' in data) this.NoAssignDefaults = data.NoAssignDefaults as boolean;
		if ('InterfaceOnly' in data) this.InterfaceOnly = data.InterfaceOnly as boolean;
		if ('MarkOptional' in data) this.MarkOptional = data.MarkOptional as boolean;
		if ('NoConstructor' in data) this.NoConstructor = data.NoConstructor as boolean;
		if ('NoToObject' in data) this.NoToObject = data.NoToObject as boolean;
		if ('NoDate' in data) this.NoDate = data.NoDate as boolean;
		if ('indents' in data) this.indents = data.indents as string[];
	}

	toObject(): { [key:string]: any } {
		const data: { [key:string]: any } = {};

		if (this.Indent) data.Indent = this.Indent;
		if (this.NoAssignDefaults) data.NoAssignDefaults = this.NoAssignDefaults;
		if (this.InterfaceOnly) data.InterfaceOnly = this.InterfaceOnly;
		if (this.MarkOptional) data.MarkOptional = this.MarkOptional;
		if (this.NoConstructor) data.NoConstructor = this.NoConstructor;
		if (this.NoToObject) data.NoToObject = this.NoToObject;
		if (this.NoDate) data.NoDate = this.NoDate;
		if (this.indents) data.indents = this.indents;
		return data;
	}
}
```

## TODO

* Use [xast](https://github.com/OneOfOne/struct2ts) to skip reflection.
* Support ES6.
* Support annoymous structs.

## License

This project is released under the [BSD 3-clause "New" or "Revised" License](https://github.com/golang/go/blob/master/LICENSE).
