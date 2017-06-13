package main

import (
	"os"

	flags "github.com/jessevdk/go-flags"
)

// GenOpts Code generator options
type GenOpts struct {
	Config flags.Filename `short:"c" long:"config" description:"Path to graphql.yaml file"`

	Package string `short:"p" long:"package" description:"Name of the package to generate"`

	Tests bool `short:"t" long:"tests" description:"Generate tests"`
}

// BoilerOpts Boilerplate options
type BoilerOpts struct {
	Force bool `short:"f" long:"force" description:"Force overwrite of existing files"`
}

// TODO: Add '@deprecated( reason: "reason" )' anotation support

// func printSource(loc *ast.Location) {
// 	str := loc.Source.Body[loc.Start:loc.End]
// 	fmt.Println(string(str))
// }

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	// genopts := GenOpts{
	// 	Config:  "graphql.yaml",
	// 	Package: "schema",
	// 	Tests:   false,
	// }
	// parser := flags.NewParser(&genopts, flags.Default)
	// boileropts := BoilerOpts{}
	// parser.AddCommand("boiler", "Creates boiler plate files",
	// 	"The boiler command creates boiler plate files, use -f to overwrite existing files",
	// 	&boileropts)
	// out, _ := parser.Parse()
	// fmt.Println(len(out), out)
	// if len(out) == 0 {
	// 	parser.WriteHelp(os.Stdout)
	// }

	// parser.WriteHelp(os.Stdout)
	// flags.Parse(&opt)
	// spew.Dump(opt)

	file := os.Args[1]

	gen, _ := newGenerator(file)
	gen.generate()
}
