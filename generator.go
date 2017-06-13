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

type Generator struct {
	Code     string
	Schema   string
	Template *template.Template
	Ast      *ast.Document
	Config   GenConfig
}

type GenConfig struct {
	Pkg        string
	ImportPath string
}

func NewGenerator(schemaFile string) (*Generator, error) {
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

	gen := &Generator{
		Schema: string(file),
		Ast:    AST,
		Config: GenConfig{
			Pkg:        "graphql",
			ImportPath: "github.com/graphql-go/graphql",
		},
	}

	gen.Template, err = template.New("main").Funcs(gen.FuncMap()).ParseGlob("language/go/*.tmpl")
	check(err)

	return gen, nil
}

type NamedDefinition interface {
	GetName() *ast.Name
	GetKind() string
}

func (gen *Generator) NamedLookup(name string) string {
	nodes := gen.Ast.Definitions

	for _, node := range nodes {
		named, ok := node.(NamedDefinition)
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

type GeneratorPass struct {
	Name string
	File string
}

var passes = []GeneratorPass{
	GeneratorPass{
		Name: "Def",
		File: "definitions.go",
	},
	GeneratorPass{
		Name: "Adp",
		File: "adapters.go",
	},
}

func (gen *Generator) Generate() {
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
