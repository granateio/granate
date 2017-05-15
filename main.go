package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"

	"github.com/go-code-gen/lib"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// TODO: Add '@deprecated( reason: "reason" )' anotation support

type GenTemplate map[string]func(interface{}, *GenConfig) string

type GraphqlGen struct {
	Code      string
	Schema    string
	Templates *template.Template
	Ast       *ast.Document
	Config    GenConfig
}

type GenConfig struct {
	Pkg        string
	ImportPath string
}

func NewGraphqlGen(schemaFile string) (*GraphqlGen, error) {
	file, err := ioutil.ReadFile(schemaFile)
	check(err)

	src := source.NewSource(&source.Source{
		Body: file,
		Name: "Schema",
	})

	AST, err := parser.Parse(parser.ParseParams{
		Source: src,
	})

	check(err)

	gen := &GraphqlGen{
		Schema: string(file),
		Ast:    AST,
		Config: GenConfig{
			Pkg:        "graphql",
			ImportPath: "github.com/graphql-go/graphql",
		},
	}

	tmpl := template.New("main")

	funcMap := lib.NewGQLTmplFuncs(tmpl)

	funcMap["cfg"] = func() GenConfig { return gen.Config }

	_, err = tmpl.Funcs(funcMap).ParseGlob("language/go/*.tmpl")
	check(err)

	gen.Templates = tmpl

	return gen, nil
}

func printSource(loc *ast.Location) {
	str := loc.Source.Body[loc.Start:loc.End]
	fmt.Println(string(str))
}

func (gen *GraphqlGen) generate() {
	nodes := gen.Ast.Definitions
	tmpl := gen.Templates

	var code bytes.Buffer

	for k, n := range nodes {
		var _ = k
		// printSource(n.GetLoc())
		err := tmpl.ExecuteTemplate(&code, n.GetKind(), n)
		check(err)
	}

	fmt.Println(code.String())

	// output, err := format.Source(code.Bytes())
	// check(err)
	// fmt.Println(string(output))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	gen, _ := NewGraphqlGen("./type.graphql")
	gen.generate()
}
