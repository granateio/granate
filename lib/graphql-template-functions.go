package lib

import (
	"bytes"
	"html/template"

	"github.com/graphql-go/graphql/language/ast"
)

type GQLTmplFuncs struct {
	Template *template.Template
}

func NewGQLTmplFuncs(tmpl *template.Template) template.FuncMap {
	tmplFunc := GQLTmplFuncs{
		Template: tmpl,
	}
	return template.FuncMap{
		"body": tmplFunc.GetBody,
		"type": tmplFunc.GetType,
		"docs": tmplFunc.GetDocs,
	}
}

func (tmpl GQLTmplFuncs) GetBody(n ast.Node) string {
	body := n.GetLoc().Source.Body
	return string(body[n.GetLoc().Start:n.GetLoc().End])
}

func (tmpl GQLTmplFuncs) GetType(t ast.Type) string {
	switch v := t.(type) {
	case *ast.Named:
		l := v.Loc
		return string(l.Source.Body[l.Start:l.End])
	case *ast.NonNull:
		var output bytes.Buffer
		l := v.Loc
		val := string(l.Source.Body[l.Start : l.End-1])
		newLoc := ast.NewLocation(l)
		newLoc.End -= 1
		newType := ParseType(val, newLoc)

		tmpl.Template.ExecuteTemplate(&output, "NonNull", map[string]ast.Type{
			"Type": newType,
		})
		return output.String()
	case *ast.List:
		var output bytes.Buffer
		l := v.Loc
		val := string(l.Source.Body[l.Start+1 : l.End-1])
		newLoc := ast.NewLocation(l)

		newLoc.End -= 1
		newLoc.Start += 1

		newType := ParseType(val, newLoc)

		tmpl.Template.ExecuteTemplate(&output, "List", map[string]ast.Type{
			"Type": newType,
		})

		return output.String()

	}
	return ""
}

func (tmpl GQLTmplFuncs) GetDocs(n ast.Node) string {

	return FindCommentBlock(n.GetLoc().Source.Body, n.GetLoc().Start)
}
