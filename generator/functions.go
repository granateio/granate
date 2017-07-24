package generator

import (
	"bytes"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/davecgh/go-spew/spew"
	"github.com/granateio/granate/generator/utils"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

func (gen *Generator) funcMap() template.FuncMap {
	return template.FuncMap{
		"cfg":           gen.getConfig,
		"graphqltype":   gen.graphqltype,
		"nativetype":    gen.nativetype,
		"nativetypepkg": gen.nativetypepkg,
		"nodes":         gen.getNodes,
		"output":        gen.getOutput,
		"root":          gen.isRootField,
		"namedtype":     gen.getNamedType,

		// Move to utils package?
		"body":         getBody,
		"desc":         getDescription,
		"kind":         getKind,
		"private":      private,
		"public":       public,
		"relay":        isRelayInterface,
		"connection":   isRelayConnection,
		"relayinput":   gen.isRelayInput,
		"relaypayload": gen.isRelayPayload,

		// Userful string functions
		"suffix": strings.HasSuffix,
		"prefix": strings.HasPrefix,

		"existfile": fileExists,
		// Placeholder functions, these functions will be replaced with a local
		// representation in each go routine for every main template
		"startfile": func() string { return "" },
		"endfile":   func() string { return "" },
		"partial":   func() string { return "" },
	}
}

func fileExists(path string) bool {
	if _, err := os.Stat("./" + path); !os.IsNotExist(err) {
		return true
	}
	return false
}

func (gen *Generator) getOutput() map[string]string {
	return gen.Config.Output
}

func isRelayInterface(iface []*ast.Named) bool {
	for _, name := range iface {
		if name.Name.Value == "Node" {
			return true
		}
	}
	return false
}

func isRelayConnection(t ast.Type) bool {
	name := getBody(t)
	return strings.HasSuffix(name, "Connection")
}

func (gen *Generator) isRelayInput(input string) bool {
	return gen.validateRelayMutationType(input, MutationInput)
}

func (gen *Generator) isRelayPayload(input string) bool {
	return gen.validateRelayMutationType(input, MutationPayload)
}

type MutationType string

const (
	MutationInput   MutationType = "Input"
	MutationPayload MutationType = "Payload"
)

func (gen *Generator) validateRelayMutationType(name string, mt MutationType) bool {
	if !strings.HasSuffix(name, string(mt)) {
		return false
	}
	srcmut := strings.ToLower(strings.TrimSuffix(name, string(mt)))
	mutnode := gen.NamedLookup("Mutation")
	if mutnode == nil {
		return false
	}
	mut := mutnode.(*ast.ObjectDefinition)
	for _, field := range mut.Fields {
		if strings.ToLower(field.Name.Value) == srcmut {
			return true
		}
	}

	return false
}

func getKind(node ast.Node) string {
	return node.GetKind()
}

func public(name string) string {
	return strings.Title(name)
}

func private(name string) string {
	index := strings.ToLower(string(name[0]))
	return index + name[1:]
}

// TODO: Load root functions from language config
func (gen *Generator) isRootField(name string) bool {
	return gen.LangConf.IsRoot(name)
}

func (gen *Generator) getNodes() astNodes {
	return gen.Nodes
}

func (gen *Generator) definition(name string) string {
	var output bytes.Buffer
	gen.Template.ExecuteTemplate(
		&output, "Graphql"+gen.NamedLookup(name).GetKind(), map[string]string{
			"Name": name,
		})
	return output.String()
}

func (gen *Generator) getConfig() interface{} {
	return gen.TmplConf
}

func getBody(n ast.Node) string {
	body := n.GetLoc().Source.Body
	return string(body[n.GetLoc().Start:n.GetLoc().End])
}

func getDescription(n ast.Node) []string {
	return utils.GetCommentBlock(n.GetLoc().Source.Body, n.GetLoc().Start)
}

func (gen *Generator) nativetypepkg(def interface{}, pkg string) string {
	return gen.def2Type(typeNative, def, pkg)
}

func (gen *Generator) nativetype(def interface{}) string {
	return gen.def2Type(typeNative, def, "")
}

func (gen *Generator) graphqltype(def interface{}) string {
	return gen.def2Type(typeGraphql, def, "")
}

func (gen *Generator) def2Type(set typeClass, def interface{}, pkg string) string {
	switch t := def.(type) {
	case *ast.Name:
		return gen.getType(set, &ast.Named{
			Kind: kinds.Named,
			Loc:  t.GetLoc(),
		}, pkg)
	case ast.Type:
		return gen.getType(set, t, pkg)
	}
	spew.Dump(def)

	// TODO: Improve error message
	log.Panicf("Unsupported type %v", def)
	return ""
}

type typeClass string

const (
	typeNative  typeClass = "Native"
	typeGraphql typeClass = "Graphql"
)

func (gen *Generator) getNamedType(t ast.Type) string {
	switch v := t.(type) {
	case *ast.Named:
		return getBody(v)
	case *ast.NonNull:
		l := v.Loc
		val := string(l.Source.Body[l.Start : l.End-1])
		newLoc := ast.NewLocation(l)
		newLoc.End--
		return gen.getNamedType(utils.ParseType(val, newLoc))
	case *ast.List:
		l := v.Loc
		val := string(l.Source.Body[l.Start+1 : l.End-1])
		newLoc := ast.NewLocation(l)

		newLoc.End--
		newLoc.Start++
		return gen.getNamedType(utils.ParseType(val, newLoc))
	}
	return ""
}

// TODO: Refactor/improve this method
func (gen *Generator) getType(typeclass typeClass, t ast.Type, pkg string) string {
	class := string(typeclass)
	switch v := t.(type) {
	case *ast.Named:
		var output bytes.Buffer
		l := v.Loc
		name := string(l.Source.Body[l.Start:l.End])

		namedType, ok := gen.LangConf.Language.Scalars[name]

		starprefix := ""
		if strings.HasPrefix(pkg, "*") {
			starprefix = "*"
			pkg = strings.TrimPrefix(pkg, "*")
		}
		if ok == true {
			if class == string(typeGraphql) {
				namedType = name
			}
			gen.Template.ExecuteTemplate(&output, class+"Named", map[string]string{
				"Name": starprefix + namedType,
			})
			return output.String()
		}

		pkgprefix := ""
		if pkg != "" {
			pkgprefix = pkg + "."
		}
		gen.Template.ExecuteTemplate(&output,
			class+gen.NamedLookup(name).GetKind(),
			map[string]string{
				"Name": pkgprefix + name,
			},
		)

		return output.String()

	case *ast.NonNull:
		var output bytes.Buffer
		l := v.Loc
		val := string(l.Source.Body[l.Start : l.End-1])
		newLoc := ast.NewLocation(l)
		newLoc.End--
		innerType := utils.ParseType(val, newLoc)

		gen.Template.ExecuteTemplate(&output, class+"NonNull", map[string]interface{}{
			"Type":    innerType,
			"Package": pkg,
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

		gen.Template.ExecuteTemplate(&output, class+"List", map[string]interface{}{
			"Type":    newType,
			"Package": pkg,
		})

		return output.String()

	}
	return ""
}
