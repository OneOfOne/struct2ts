#!/bin/bash -Xe

go run github.com/OneOfOne/struct2ts/cmd/struct2ts github.com/OneOfOne/struct2ts/testdata/testmodel1.Struct1
go run github.com/OneOfOne/struct2ts/cmd/struct2ts github.com/OneOfOne/struct2ts/testdata/testmodel1.Struct2
go run github.com/OneOfOne/struct2ts/cmd/struct2ts github.com/OneOfOne/struct2ts/testdata/testmodel1.Struct3
