package testmodel1

import (
	"fmt"
	"io"
)

type Struct1 struct {
	HasNoCustomInterface int
}

type Struct2 struct {
	TotallyDoesHaveOneAsAPointer int
}

func (s *Struct2) RenderCustomTypescript(w io.Writer) (err error) {
	fmt.Fprint(w, "// Custom Output!!!")
	return nil
}

type Struct3 struct {
	TotallyDoesHaveOne int
}

func (s Struct3) RenderCustomTypescript(w io.Writer) (err error) {
	fmt.Fprint(w, "// Custom Output!!!")
	return nil
}
