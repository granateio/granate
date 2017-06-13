package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

type generator struct {
	Code     string
	Schema   string
	Template *template.Template
	Ast      *ast.Document
	Config   genConfig
}

type genConfig struct {
	Pkg        string
	ImportPath string
}

func newGenerator(schemaFile string) (*generator, error) {
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

	gen := &generator{
		Schema: string(file),
		Ast:    AST,
		Config: genConfig{
			Pkg:        "graphql",
			ImportPath: "github.com/graphql-go/graphql",
		},
	}

	gen.Template, err = template.New("main").Funcs(gen.funcMap()).ParseGlob("language/go/*.tmpl")
	check(err)

	return gen, nil
}

type namedDefinition interface {
	GetName() *ast.Name
	GetKind() string
}

func (gen *generator) NamedLookup(name string) string {
	nodes := gen.Ast.Definitions

	for _, node := range nodes {
		named, ok := node.(namedDefinition)
		if ok == false {
			continue
		}
		if named.GetName().Value == name {
			return named.GetKind()
		}
	}

	log.Fatalf("Type with name '%s' is not defined", name)
	return ""
}

type generatorPass struct {
	Name string
	File string
}

var passes = []generatorPass{
	generatorPass{
		Name: "Def",
		File: "definitions.go",
	},
	generatorPass{
		Name: "Adp",
		File: "adapters.go",
	},
}

func (gen *generator) generate() {
	nodes := gen.Ast.Definitions
	tmpl := gen.Template

	for _, pass := range passes {
		var code bytes.Buffer
		err := tmpl.ExecuteTemplate(&code, "Header", nil)
		_ = err
		for _, n := range nodes {
			err := tmpl.ExecuteTemplate(&code, pass.Name+"_"+n.GetKind(), n)
			_ = err
			// check(err)
		}
		fmt.Println(code.String())
	}

}
