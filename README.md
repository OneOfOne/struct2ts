# go2ts

An extremely simple and powerful Go struct to Typescript interface generator.

Inspired by [https://github.com/OneOfOne/struct2ts](https://github.com/OneOfOne/struct2ts).

## Install

    go get

## Example

Input:

```go
type ComplexStruct struct {
	S           string       `json:"s,omitempty"`
	I           int          `json:"i,omitempty"`
	F           float64
}

func main() {
	s := go2ts.New()
	s.Add(machine.Description{})
	s.Render(os.Stdout)
}
```

Output:

```ts
export interface ComplexStruct {
  s?: string;
  i?: number;
  F: number;
}
```
