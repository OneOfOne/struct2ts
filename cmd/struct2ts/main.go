package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	s2ts "github.com/OneOfOne/struct2ts"
	"github.com/alecthomas/template"
	"golang.org/x/tools/imports"
	KP "gopkg.in/alecthomas/kingpin.v2"
)

const version = "v0.0.1"

var (
	opts     s2ts.Options
	types    []string
	keepTemp bool

	tmpl = template.Must(template.New("").Parse(fileTmpl))
)

func init() {
	KP.Flag("indent", "output indentation").Default("\t").StringVar(&opts.Indent)
	KP.Flag("mark-optional-fields", "add `?` to fields with omitempty").Short('m').BoolVar(&opts.MarkOptional)
	KP.Flag("no-ctor", "don't generate a constructor").Short('C').BoolVar(&opts.NoConstructor)
	KP.Flag("no-toObject", "don't generate a Class.toObject() method").Short('T').BoolVar(&opts.NoToObject)
	KP.Flag("no-date", "don't automatically handle time.Unix () <-> JS Date()").Short('D').BoolVar(&opts.NoDate)
	KP.Flag("no-default-values", "don't assign default/zero values in the ctor").Short('N').BoolVar(&opts.NoAssignDefaults)
	KP.Flag("interface", "only generate an interface (disables all the other options)").Short('i').BoolVar(&opts.InterfaceOnly)

	KP.Flag("keep-temp", "keep the generated tmp file").Short('k').BoolVar(&keepTemp)

	KP.Arg("pkg.struct", "list of structs to convert").Required().HintOptions("github.com/you/auth/users.User").StringsVar(&types)
}

type M = map[string]interface{}

func main() {
	log.SetFlags(log.Lshortfile)
	KP.HelpFlag.Short('h')
	KP.Version(version).VersionFlag.Short('V')
	KP.Parse()

	src, err := render()
	if err != nil {
		log.Fatal(err)
	}

	if src, err = imports.Process("s2ts_gen.go", src, nil); err != nil {
		log.Fatal(err)
	}

	f, err := ioutil.TempFile("", "s2ts_gen_*.go")
	if err != nil {
		log.Fatal(err)
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

	log.Printf("executing go run %s", f.Name())
	cmd := exec.Command("go", "run", f.Name())

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err = cmd.Run(); err != nil {
		log.Panic(err)
	}
}

func render() ([]byte, error) {
	var (
		buf     bytes.Buffer
		imports []string
		ttypes  = types[:0]
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
		ttypes = append(ttypes, t)
	}
	err := tmpl.Execute(&buf, M{
		"opts":    opts,
		"imports": imports,
		"types":   ttypes,
	})
	return buf.Bytes(), err
}

const fileTmpl = `package main

import (
	"os"

	{{ range $_, $imp := .imports }}
	"{{$imp}}"{{ end }}

	s2ts "github.com/OneOfOne/struct2ts"

)

func main() {
	s := s2ts.New(&s2ts.Options{
		Indent: "{{ .opts.Indent }}",

		NoAssignDefaults: {{ .opts.NoAssignDefaults }},
		InterfaceOnly:    {{ .opts.InterfaceOnly    }},

		NoConstructor: {{ .opts.NoConstructor }},
		MarkOptional:  {{ .opts.MarkOptional  }},
		NoToObject:    {{ .opts.NoToObject    }},
		NoDate:        {{ .opts.NoDate        }},
	})

	{{ range $_, $t := .types }}
	s.Add({{$t}}{}){{ end }}

	if err := s.RenderTo(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
`
