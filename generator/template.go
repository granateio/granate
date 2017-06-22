package generator

import (
	"bytes"
	"html/template"
	"log"
	"strings"

	"github.com/granate/generator/utils"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

func (gen *Generator) funcMap() template.FuncMap {
	return template.FuncMap{
		"def2native":  gen.def2Native,
		"def2graphql": gen.def2Graphql,
		"desc":        gen.description,
		"cfg":         gen.getConfig,
		"public":      gen.public,
		"body":        gen.getBody,
		"core":        gen.core,
	}
}

// TODO: Load core functions from language config
func (gen *Generator) core(name string) bool {
	switch name {
	case
		"Query",
		"Mutation",
		"Subscription":
		return true
	}
	return false
}

func (gen *Generator) definition(name string) string {
	var output bytes.Buffer
	gen.Template.ExecuteTemplate(
		&output, "Graphql"+gen.NamedLookup(name), map[string]string{
			"Name": name,
		})
	return output.String()
}

func (gen *Generator) getConfig() interface{} {
	return gen.TmplConf
}

func (gen Generator) public(name string) string {
	return strings.Title(name)
}

func (gen Generator) getBody(n ast.Node) string {
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

func (gen Generator) description(n ast.Node) []string {
	return utils.GetCommentBlock(n.GetLoc().Source.Body, n.GetLoc().Start)
}

// TODO: Discuss naming conventions for:
//	Def2Native, Def2Graphql and Def2Type

func (gen *Generator) def2Native(def interface{}) string {
	return gen.def2Type(typeNative, def)
}

func (gen *Generator) def2Graphql(def interface{}) string {
	return gen.def2Type(typeGraphql, def)
}

func (gen *Generator) def2Type(set typeSet, def interface{}) string {
	switch t := def.(type) {
	case *ast.Name:
		return gen.getType(set, &ast.Named{
			Kind: kinds.Named,
			Loc:  t.GetLoc(),
		})
	case ast.Type:
		return gen.getType(set, t)
	}

	// TODO: Improve error message
	log.Panicf("Unsupported type %v", def)
	return ""
}

type typeSet string

const (
	typeNative  typeSet = "Native"
	typeGraphql typeSet = "Graphql"
)

// TODO: Refactor/improve this method
func (gen *Generator) getType(typeset typeSet, t ast.Type) string {
	set := string(typeset)
	switch v := t.(type) {
	case *ast.Named:
		var output bytes.Buffer
		l := v.Loc
		name := string(l.Source.Body[l.Start:l.End])
		if namedGraphqlType(name) == true {
			gen.Template.ExecuteTemplate(&output, set+"Named", map[string]string{
				// TODO: Fetch the correct name for native and graphql types
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
		newLoc.End--
		innerType := utils.ParseType(val, newLoc)

		gen.Template.ExecuteTemplate(&output, set+"NonNull", map[string]ast.Type{
			"Type": innerType,
		})
		return output.String()
	case *ast.List:
		var output bytes.Buffer
		l := v.Loc
		val := string(l.Source.Body[l.Start+1 : l.End-1])
		newLoc := ast.NewLocation(l)

		newLoc.End--
		newLoc.Start++

		newType := utils.ParseType(val, newLoc)

		gen.Template.ExecuteTemplate(&output, set+"List", map[string]ast.Type{
			"Type": newType,
		})

		return output.String()

	}
	return ""
}
