package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	yaml "gopkg.in/yaml.v2"

	"github.com/davecgh/go-spew/spew"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

type generator struct {
	Code     string
	Schema   string
	Template *template.Template
	Ast      *ast.Document
	Generate generatorConfig
	// TODO: Remove Config issue: #1
	Config genConfig
}

// TODO genConfig and generatorConfig got to similar names
type genConfig struct {
	Pkg        string
	ImportPath string
}

type generatorConfig struct {
	// TODO Support a globbing system
	Schemas  []string
	Language string
	Package  string
}

// func newGenerator(schemaFile string) (*generator, error) {
func newGenerator(config string) (*generator, error) {

	confFile, err := ioutil.ReadFile(config)
	check(err)

	genCfg := generatorConfig{}
	err = yaml.Unmarshal(confFile, &genCfg)
	check(err)

	spew.Dump(&genCfg)

	// Combine all .graphql files into one schema
	var schema bytes.Buffer
	for _, scm := range genCfg.Schemas {
		file, err := ioutil.ReadFile(scm)
		check(err)
		schema.Write(file)
	}

	// Create the package directory
	// Ignore error for now
	err = os.Mkdir(genCfg.Package, 0766)

	// log.Fatal(schema.String())

	src := source.NewSource(&source.Source{
		Body: schema.Bytes(),
		Name: "Schema",
	})

	AST, err := parser.Parse(parser.ParseParams{
		Source: src,
	})

	check(err)

	gen := &generator{
		Schema: schema.String(),
		Ast:    AST,
		Config: genConfig{
			Pkg:        "graphql",
			ImportPath: "github.com/graphql-go/graphql",
		},
		Generate: genCfg,
	}

	gen.Template, err = template.New("main").
		Funcs(gen.funcMap()).
		ParseGlob("language/go/*.tmpl")

	check(err)

	return gen, nil
}

type namedDefinition interface {
	GetName() *ast.Name
	GetKind() string
}

// TODO: Find a better name for the NamedLookup function
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

// TODO: Should rethink the generator pass system issue: #4
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

		// Code output
		filename := gen.Generate.Package + "/" + pass.File
		fmt.Println(filename)

		// TODO: Read the fmt command from config
		cmd := exec.Command("gofmt")
		stdin, err := cmd.StdinPipe()
		check(err)

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, code.String())
		}()

		out, err := cmd.CombinedOutput()
		// Format code here
		err = ioutil.WriteFile(filename, out, 0644)
		check(err)
	}

}
