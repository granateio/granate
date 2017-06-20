package main

import (
	"fmt"

	"github.com/graphql-go-gen/schema"
)

func main() {
	schema.Provider.Query = Query{}
	fmt.Println("vim-go")
}
