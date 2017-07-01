package generator

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"

	yaml "gopkg.in/yaml.v2"

	"github.com/granate/generator/utils"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// Generator represents the code generator main object
type Generator struct {
	Code     string
	Schema   string
	Template *template.Template
	Ast      *ast.Document
	Config   ProjectConfig
	LangConf LanguageConfig

	TmplConf map[string]string
}

// ProjectConfig contains the granate.yaml information
type ProjectConfig struct {
	// TODO Support a globbing system
	Schemas  []string
	Language string
	Package  string
}

// LanguageConfig defines the language specific
// implementation information
type LanguageConfig struct {

	// Language specific syntax config
	Language struct {
		Scalars map[string]string
		Root    []string
	}

	// This is passed to the generators Cfg variable
	Config map[string]string

	// Prefix for the templates and the output filename
	Passes struct {
		Prefix   string
		filename string
	}

	// Extra templates for each pass without prefix
	Templates []string

	// Program/command used for formatting the output code
	Formatter struct {
		CMD  string
		Args []string
	}
}

func (lang LanguageConfig) IsRoot(val string) bool {
	for _, root := range lang.Language.Root {
		if root == val {
			return true
		}
	}

	return false
}

// New creates a new Generator instance
func New(config string) (*Generator, error) {

	confFile, err := ioutil.ReadFile(config)
	check(err)

	genCfg := ProjectConfig{}
	err = yaml.Unmarshal(confFile, &genCfg)
	check(err)

	// Combine all .graphql files into one schema
	var schema bytes.Buffer
	for _, scm := range genCfg.Schemas {
		file, err := ioutil.ReadFile(scm)
		check(err)
		schema.Write(file)
	}

	// Create the generated package directory
	// Ignore error for now
	err = os.Mkdir(genCfg.Package, 0766)

	src := source.NewSource(&source.Source{
		Body: schema.Bytes(),
		Name: "Schema",
	})

	AST, err := parser.Parse(parser.ParseParams{
		Source: src,
	})

	check(err)

	gopath := os.Getenv("GOPATH")
	projectpath := gopath + "/src/github.com/granate/"
	langpath := projectpath + "language/" + genCfg.Language + "/"

	langConfigFile, err := ioutil.ReadFile(langpath + "config.yaml")
	check(err)

	langConfig := LanguageConfig{}
	err = yaml.Unmarshal(langConfigFile, &langConfig)
	check(err)

	gen := &Generator{
		Schema:   schema.String(),
		Ast:      AST,
		TmplConf: langConfig.Config,
		Config:   genCfg,
		LangConf: langConfig,
	}

	gen.Template, err = template.New("main").
		Funcs(gen.funcMap()).
		ParseGlob(langpath + "*.tmpl")

	check(err)

	return gen, nil
}

type namedDefinition interface {
	GetName() *ast.Name
	GetKind() string
}

// TODO: Find a better name for the NamedLookup function
func (gen *Generator) NamedLookup(name string) string {
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

func (gen generatorPass) template(name string) string {
	return gen.Name + "_" + name
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

// Generate starts the code generation process
func (gen *Generator) Generate() {
	nodes := gen.Ast.Definitions
	tmpl := gen.Template
	var lines int

	var wait sync.WaitGroup

	for _, pass := range passes {
		wait.Add(1)

		go func(pass generatorPass) {
			defer wait.Done()

			var code bytes.Buffer
			err := tmpl.ExecuteTemplate(&code, pass.template("Header"), nil)
			_ = err
			for _, n := range nodes {
				err := tmpl.ExecuteTemplate(&code, pass.template(n.GetKind()), n)
				_ = err
			}

			// Code output
			filename := gen.Config.Package + "/" + pass.File
			fmt.Println(filename)

			debug := os.Getenv("GRANATE_DEBUG")
			if debug == "true" {
				fmt.Println(code.String())
			}

			// TODO: Read the fmt command from config
			cmd := exec.Command("gofmt")
			stdin, err := cmd.StdinPipe()
			check(err)

			go func() {
				defer stdin.Close()
				io.WriteString(stdin, code.String())
			}()

			out, err := cmd.CombinedOutput()

			ln, _ := utils.LineCounter(bytes.NewReader(out))
			lines += ln

			err = ioutil.WriteFile(filename, out, 0644)
			check(err)
		}(pass)
	}

	wait.Wait()

	fmt.Printf("Generated %d lines of code\n", lines)

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
