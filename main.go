package main

import (
	"os"

	"github.com/granateio/granate/generator"
	flags "github.com/jessevdk/go-flags"
)

// Flags Code generator options
type Flags struct {
	Config string `short:"c" long:"config" description:"Path to <config>.yaml file"`
	Help   bool   `short:"h" long:"help" description:"Show available options"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	params := Flags{
		Config: "granate.yaml",
		Help:   false,
	}

	parser := flags.NewParser(&params, flags.Default^flags.HelpFlag)
	_, err := parser.Parse()
	check(err)

	if params.Help == true {
		parser.WriteHelp(os.Stdout)
		return
	}

	file := params.Config

	gen, _ := generator.New(file)
	gen.Generate()
}
