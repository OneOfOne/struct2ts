package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"text/template"

	"github.com/OneOfOne/struct2ts"
	"golang.org/x/tools/imports"
	KP "gopkg.in/alecthomas/kingpin.v2"
)

const version = "v1.0.0"

var (
	opts  struct2ts.Options
	types []string

	outFile string

	srcOnly bool
	pkgName string

	keepTemp bool

	tmpl = template.Must(template.New("").Parse(fileTmpl))
)

func init() {
	KP.Flag("indent", "Output indentation.").Default("\t").StringVar(&opts.Indent)
	KP.Flag("mark-optional-fields", "Add `?` to fields with omitempty.").Short('m').BoolVar(&opts.MarkOptional)
	KP.Flag("es6", "generate es6 code").Short('6').BoolVar(&opts.ES6)
	KP.Flag("no-ctor", "Don't generate a ctor.").Short('C').BoolVar(&opts.NoConstructor)
	KP.Flag("no-toObject", "Don't generate a Class.toObject() method.").Short('T').BoolVar(&opts.NoToObject)
	KP.Flag("no-exports", "Don't automatically export the generated types.").Short('E').BoolVar(&opts.NoExports)
	KP.Flag("no-date", "Don't automatically handle time.Unix () <-> JS Date().").Short('D').BoolVar(&opts.NoDate)
	KP.Flag("no-helpers", "Don't output the helpers.").Short('H').BoolVar(&opts.NoHelpers)
	KP.Flag("no-default-values", "Don't assign default/zero values in the ctor.").Short('N').BoolVar(&opts.NoAssignDefaults)
	KP.Flag("no-capitalize", "Don't capitalize TS class names.").Short('C').BoolVar(&opts.NoCapitalize)
	KP.Flag("interface", "Only generate an interface (disables all the other options).").Short('i').BoolVar(&opts.InterfaceOnly)

	KP.Flag("src-only", "Only output the Go code (helpful if you want to edit it yourself).").Short('s').BoolVar(&srcOnly)
	KP.Flag("package-name", "the package name to use if --src-only is set.").
		Default("main").Short('p').StringVar(&pkgName)

	KP.Flag("keep-temp", "Keep the generated Go file, ignored if --src-only is set.").Short('k').BoolVar(&keepTemp)

	KP.Flag("out", "Write the output to a file instead of stdout.").Short('o').Default("-").StringVar(&outFile)

	KP.Arg("pkg.struct", "List of structs to convert (github.com/you/auth/users.User, users.User or users.User:AliasUser).").
		StringsVar(&types)

}

type M map[string]interface{}

func main() {
	log.SetFlags(log.Lshortfile)
	KP.HelpFlag.Short('h')
	KP.Version(version).VersionFlag.Short('V')
	KP.Parse()

	out := os.Stdout

	if outFile != "-" && outFile != "/dev/stdout" {
		of, err := os.Create(outFile)
		if err != nil {
			log.Panic(err)
		}
		defer of.Close()
		out = of
	}

	src, err := render()
	if err != nil {
		log.Panic(err)
	}

	if src, err = imports.Process("s2ts_gen.go", src, nil); err != nil {
		log.Panic(err)
	}

	if srcOnly {
		out.Write(src)
		return
	}

	f, err := tempFile()
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		if recover() == nil && !keepTemp {
			os.Remove(f.Name())
		}
	}()

	_, err = f.Write(src)
	f.Close()

	if err != nil {
		log.Panic(err)
	}

	log.Printf("executing: go run %s", f.Name())
	cmd := exec.Command("go", "run", f.Name())

	cmd.Stderr = os.Stderr
	cmd.Stdout = out

	if err = cmd.Run(); err != nil {
		log.Panic(err)
	}
}

func render() ([]byte, error) {
	var (
		buf            bytes.Buffer
		imports        []string
		ttypes         = types[:0]
		typesWithNames [][2]string
	)

	for _, t := range types {
		idx, dotIdx := strings.LastIndexByte(t, '/'), strings.LastIndexByte(t, '.')
		if dotIdx == -1 {
			log.Printf("%s is an invalid import.", t)
			continue
		}
		if idx > -1 {
			imports = append(imports, t[:dotIdx])
			t = t[idx+1:]
		}
		if idx := strings.LastIndexByte(t, ':'); idx != -1 {
			typesWithNames = append(typesWithNames, [2]string{t[:idx], t[idx+1:]})
		} else {
			ttypes = append(ttypes, t)
		}
	}

	err := tmpl.Execute(&buf, M{
		"pkgName":        pkgName,
		"cmd":            strings.Join(os.Args[1:], " "),
		"opts":           opts,
		"imports":        imports,
		"types":          ttypes,
		"typesWithNames": typesWithNames,
	})

	return buf.Bytes(), err
}

func tempFile() (f *os.File, err error) {
	// if this somehow conflicts, god really hates us.
	fpath := filepath.Join(os.TempDir(), fmt.Sprintf("s2ts_gen_%d_%d.go", time.Now().UnixNano(), rand.Int63()))
	return os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
}

const fileTmpl = `// this file was automatically generated using struct2ts {{.cmd}}
package {{.pkgName}}

import (
	"flag"
	"log"
	"io"
	"os"

	"github.com/OneOfOne/struct2ts"
	{{ range $_, $imp := .imports }}"{{$imp}}"{{ end }}
)
{{- if eq .pkgName "main" }}
func main() {
	log.SetFlags(log.Lshortfile)

	var (
		out = flag.String("o", "-", "output")
		f = os.Stdout
		err error
	)

	flag.Parse()
	if *out != "-" {
		if f, err = os.OpenFile(*out, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644); err != nil {
			panic(err)
		}
		defer f.Close()
	}
	if err = runStruct2TS(f); err != nil {
		panic(err)
	}
}
{{- end }}

func runStruct2TS(w io.Writer) error {
	s := struct2ts.New(&struct2ts.Options{
		Indent: "{{ .opts.Indent }}",

		NoAssignDefaults: {{ .opts.NoAssignDefaults }},
		InterfaceOnly:    {{ .opts.InterfaceOnly    }},

		NoConstructor: {{ .opts.NoConstructor }},
		NoCapitalize:  {{ .opts.NoCapitalize }},
		MarkOptional:  {{ .opts.MarkOptional  }},
		NoToObject:    {{ .opts.NoToObject    }},
		NoExports:     {{ .opts.NoExports        }},
		NoHelpers:     {{ .opts.NoHelpers        }},
		NoDate:        {{ .opts.NoDate        }},

		ES6:           {{ .opts.ES6 }},
	})

	{{ range $_, $t := .types }}
	s.Add({{$t}}{})
	{{- end }}
	{{ range $_, $t := .typesWithNames }}
	s.AddWithName({{index $t 0}}{}, "{{index $t 1}}")
	{{- end }}

	io.WriteString(w, "// this file was automatically generated, DO NOT EDIT\n")
	return s.RenderTo(w)
}
`
