package main

import (
	"github.com/granate/schema"
)

func main() {
	schema.Provider.Query = Query{
		User: users[1],
	}

	schema.Provider.Mutation = Mutation{}

	schema.Init()
}
