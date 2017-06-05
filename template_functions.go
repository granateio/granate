package main

import (
	"bytes"
	"html/template"
	"log"
	"strings"

	"github.com/base-dev/graphql/language/kinds"
	"github.com/graphql-go-gen/lib"
	"github.com/graphql-go/graphql/language/ast"
)

func (gen *Generator) FuncMap() template.FuncMap {
	return template.FuncMap{
		"def2native":  gen.Def2Native,
		"def2graphql": gen.Def2Graphql,
		"desc":        gen.Description,
		"cfg":         gen.GetConfig,
		"public":      gen.Public,
		"body":        gen.GetBody,
	}
}

func (gen *Generator) Definition(name string) string {
	var output bytes.Buffer
	gen.Template.ExecuteTemplate(
		&output, "Graphql"+gen.NamedLookup(name), map[string]string{
			"Name": name,
		})
	return output.String()
}

func (gen *Generator) GetConfig() interface{} {
	return gen.Config
}

func (gen Generator) Public(name string) string {
	return strings.Title(name)
}

func (gen Generator) GetBody(n ast.Node) string {
	body := n.GetLoc().Source.Body
	return string(body[n.GetLoc().Start:n.GetLoc().End])
}

// TODO: Move this out to the language config (language/<lang>/config.yaml)
var typemap = map[string]string{
	"String":  "string",
	"Int":     "int",
	"Float":   "float",
	"Boolean": "bool",
	"ID":      "string",
}

// TODO: Remove this
func namedGraphqlType(name string) bool {
	switch name {
	case
		"String",
		"Int",
		"Float",
		"Boolean",
		"ID":
		return true
	}
	return false
}

func (gen Generator) Description(n ast.Node) []string {
	return lib.FindCommentBlock(n.GetLoc().Source.Body, n.GetLoc().Start)
}

// TODO: Discuss naming conventions for:
//	Def2Native, Def2Graphql and Def2Type

func (gen *Generator) Def2Native(def interface{}) string {
	return gen.Def2Type(TypeNative, def)
}

func (gen *Generator) Def2Graphql(def interface{}) string {
	return gen.Def2Type(TypeGraphql, def)
}

func (gen *Generator) Def2Type(set TypeSet, def interface{}) string {
	switch t := def.(type) {
	case *ast.Name:
		return gen.GetType(set, &ast.Named{
			Kind: kinds.Named,
			Loc:  t.GetLoc(),
		})
	case ast.Type:
		return gen.GetType(set, t)
	}

	// TODO: Improve error message
	log.Panicf("Unsupported type %v", def)
	return ""
}

type TypeSet string

const (
	TypeNative  TypeSet = "Native"
	TypeGraphql TypeSet = "Graphql"
)

// TODO: Refactor/improve this method
func (gen *Generator) GetType(typeset TypeSet, t ast.Type) string {
	set := string(typeset)
	switch v := t.(type) {
	case *ast.Named:
		var output bytes.Buffer
		l := v.Loc
		name := string(l.Source.Body[l.Start:l.End])
		if namedGraphqlType(name) == true {
			gen.Template.ExecuteTemplate(&output, set+"Named", map[string]string{
				"Name": typemap[name],
			})
			return output.String()
		}
		gen.Template.ExecuteTemplate(&output, set+gen.NamedLookup(name), map[string]string{
			"Name": name,
		})
		return output.String()
	case *ast.NonNull:
		var output bytes.Buffer
		l := v.Loc
		val := string(l.Source.Body[l.Start : l.End-1])
		newLoc := ast.NewLocation(l)
		newLoc.End -= 1
		innerType := lib.ParseType(val, newLoc)

		gen.Template.ExecuteTemplate(&output, set+"NonNull", map[string]ast.Type{
			"Type": innerType,
		})
		return output.String()
	case *ast.List:
		var output bytes.Buffer
		l := v.Loc
		val := string(l.Source.Body[l.Start+1 : l.End-1])
		newLoc := ast.NewLocation(l)

		newLoc.End -= 1
		newLoc.Start += 1

		newType := lib.ParseType(val, newLoc)

		gen.Template.ExecuteTemplate(&output, set+"List", map[string]ast.Type{
			"Type": newType,
		})

		return output.String()

	}
	return ""
}
