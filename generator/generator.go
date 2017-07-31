package generator

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"text/template"

	yaml "gopkg.in/yaml.v2"

	"github.com/granateio/granate/generator/utils"
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
	Nodes    astNodes

	TmplConf map[string]string
}

// ProjectConfig contains the granate.yaml information
type ProjectConfig struct {
	// TODO Support a globbing system
	Schemas  []string
	Language string
	Output   map[string]string
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

	// Main templates, each template in this list is executed in it's own
	// go routine
	Templates []string

	// Program/command used for formatting the output code
	Formatter struct {
		CMD  string
		Args []string
	}
}

type OutputFileBuffer struct {
	Path   string
	Buffer *bytes.Buffer
}

func (out *OutputFileBuffer) GetBuffer() *bytes.Buffer {
	return out.Buffer
}

type TemplateFileFuncs struct {
	BufferStack   *utils.Lifo
	SwapBuffer    *utils.SwapBuffer
	LocalTemplate *template.Template
	linenumber    int
}

func (tmpl *TemplateFileFuncs) LineNumbers() int {
	return tmpl.linenumber
}

func (tmpl *TemplateFileFuncs) Start(path string) string {
	// Push current buffer on the stack
	tmpl.BufferStack.Push(tmpl.SwapBuffer.GetBuffer())

	// Create a new OpaqueBuffer
	output := &OutputFileBuffer{
		Path:   path,
		Buffer: &bytes.Buffer{},
	}

	tmpl.SwapBuffer.SetBuffer(output)

	return ""
}

func (tmpl *TemplateFileFuncs) End() string {

	output, ok := tmpl.SwapBuffer.GetBuffer().(*OutputFileBuffer)

	if ok == false {
		panic("GetBuffer() does not return a pointer to OutputFileBuffer")
	}

	if output.Path == "" {
		return ""
	}

	// fmt.Println(output.GetBuffer().String())
	// Unnecessary:
	// tmpl.FileBuffers = append(tmpl.FileBuffers, output)

	dir := path.Join(".", path.Dir(output.Path))
	err := os.MkdirAll(dir, os.ModePerm)

	if err != nil {
		fmt.Println(err)
	}

	// TODO: Read the fmt command from config
	// cmd := exec.Command("gofmt")
	cmd := exec.Command("goimports")
	stdin, err := cmd.StdinPipe()
	check(err)

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, output.GetBuffer().String())
	}()

	out, err := cmd.CombinedOutput()

	ln, _ := utils.LineCounter(bytes.NewReader(out))
	tmpl.linenumber += ln

	err = ioutil.WriteFile(output.Path, out, 0644)
	// err = ioutil.WriteFile(output.Path, output.GetBuffer().Bytes(), 0644)

	check(err)

	prevBuffer, ok := tmpl.BufferStack.Pop().(utils.OpaqueBytesBuffer)
	if ok == false {
		panic("Found wrong type in BufferStack")
	}

	tmpl.SwapBuffer.SetBuffer(prevBuffer)

	return ""
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
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	genCfg := ProjectConfig{}
	err = yaml.Unmarshal(confFile, &genCfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Combine all .graphql files into one schema
	var schema bytes.Buffer
	for _, scm := range genCfg.Schemas {
		file, err := ioutil.ReadFile(scm)
		check(err)
		schema.Write(file)
	}

	// Create the generated package directory
	// Ignore error for now
	// err = os.Mkdir(genCfg.Package, 0766)

	src := source.NewSource(&source.Source{
		Body: schema.Bytes(),
		Name: "Schema",
	})

	AST, err := parser.Parse(parser.ParseParams{
		Source: src,
	})

	check(err)

	gopath := os.Getenv("GOPATH")
	projectpath := gopath + "/src/github.com/granateio/granate/"
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

	// gen.Nodes.Connection = make(map[string]ast.Node)

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
func (gen *Generator) NamedLookup(name string) ast.Node {
	return NodeByName(gen.Nodes.Definition, name)
}

// TODO: Much the same as the NamedLookup function
func NodeByName(nodes []ast.Node, name string) ast.Node {
	for _, node := range nodes {
		named, ok := node.(namedDefinition)
		if ok == false {
			continue
		}
		if named.GetName().Value == name {
			return node
		}
	}

	log.Fatalf("Type with name '%s' is not defined", name)
	return nil
}

type generatorPass struct {
	Name string
	File string
}

func (gen generatorPass) template(name string) string {
	return gen.Name + "_" + name
}

type astNodes struct {
	Root       []ast.Node
	Definition []ast.Node
	Object     []ast.Node
	Relay      []ast.Node
}

type ConnectionDefinition struct {
	Name     *ast.Name
	Loc      *ast.Location
	NodeType ast.Node
}

func (con ConnectionDefinition) GetKind() string {
	return "ConnectionDefinition"
}

func (con ConnectionDefinition) GetLoc() *ast.Location {
	return con.Loc
}

func (con ConnectionDefinition) GetName() *ast.Name {
	return con.Name
}

// Generate starts the code generation process
func (gen *Generator) Generate() {
	definitions := gen.Ast.Definitions

	tmpl := gen.Template
	mainTemplates := gen.LangConf.Templates

	var wait sync.WaitGroup
	var nodes astNodes
	connections := make(map[string]bool)

	// Gather usefull definitions
	for _, def := range definitions {
		namedef, ok := def.(namedDefinition)

		if ok == false {
			continue
		}

		nodes.Definition = append(nodes.Definition, def)

		if gen.LangConf.IsRoot(namedef.GetName().Value) {
			nodes.Root = append(nodes.Root, def)
		}

		objectDef, ok := def.(*ast.ObjectDefinition)
		if ok == false {
			continue
		}

		nodes.Object = append(nodes.Object, def)

		// Find and add relay connections
		for _, connection := range objectDef.Fields {
			conloc := connection.Type.GetLoc()
			contype := string(conloc.Source.Body[conloc.Start:conloc.End])
			if strings.HasSuffix(contype, "Connection") {
				// if _, ok := nodes.Connection[contype]; ok == true {
				if _, ok := connections[contype]; ok == true {
					continue
				}
				con := ConnectionDefinition{
					Name: ast.NewName(&ast.Name{
						Value: contype,
						Loc:   conloc,
					}),
					Loc:      conloc,
					NodeType: NodeByName(gen.Ast.Definitions, strings.TrimSuffix(contype, "Connection")),
				}
				nodes.Definition = append(nodes.Definition, con)
				connections[contype] = true
			}
		}

		for _, iface := range objectDef.Interfaces {
			body := string(iface.Loc.Source.Body)
			name := body[iface.Loc.Start:iface.Loc.End]
			if name == "Node" {
				nodes.Relay = append(nodes.Relay, def)
			}
		}
	}

	gen.Nodes = nodes

	linecounter := make(chan int)
	quit := make(chan bool)

	go func(quit chan bool, counter chan int) {
		sum := 0
		for {
			select {
			case number := <-counter:
				sum += number
			case <-quit:
				fmt.Println("Generated", sum, "lines of code")
				return
			}
		}
	}(quit, linecounter)

	for _, mainTmpl := range mainTemplates {
		wait.Add(1)

		go func(mainTmpl string, counter chan int) {
			defer wait.Done()

			localTemplate, err := tmpl.Clone()
			if err != nil {
				panic(err)
			}

			codebuffer := &utils.SwapBuffer{}
			codebuffer.SetBuffer(&OutputFileBuffer{
				Buffer: &bytes.Buffer{},
			})

			localFileFuncs := TemplateFileFuncs{
				BufferStack:   &utils.Lifo{},
				SwapBuffer:    codebuffer,
				LocalTemplate: localTemplate,
			}

			partialfunc := func(name string, data interface{}) string {
				localbuffer := bytes.Buffer{}
				localTemplate.ExecuteTemplate(&localbuffer, name, data)
				return localbuffer.String()
			}

			fileFuncsMap := template.FuncMap{
				"startfile": localFileFuncs.Start,
				"endfile":   localFileFuncs.End,
				"partial":   partialfunc,
			}

			localTemplate = localTemplate.Funcs(fileFuncsMap)

			err = localTemplate.ExecuteTemplate(codebuffer, mainTmpl, nil)
			if err != nil {
				panic(err)
			}

			counter <- localFileFuncs.LineNumbers()

		}(mainTmpl, linecounter)
	}

	wait.Wait()

	quit <- true

	// fmt.Printf("Generated %d lines of code\n", lines)

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
