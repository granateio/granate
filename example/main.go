//go:generate granate

package main

import (
	"github.com/granateio/granate/example/schema"
)

func main() {
	schema.Provider.Query = Query{
		User: users[1],
	}

	schema.Provider.Mutation = Mutation{}

	schema.Init()
}
