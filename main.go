package main

import (
	"fmt"

	"github.com/graphql-go/graphql/language/ast"
)

// TODO: Add '@deprecated( reason: "reason" )' anotation support

// type GenTemplate map[string]func(interface{}, *GenConfig) string

func printSource(loc *ast.Location) {
	str := loc.Source.Body[loc.Start:loc.End]
	fmt.Println(string(str))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	gen, _ := NewGenerator("./todo.graphql")
	gen.Generate()
}
