package main

import (
	"fmt"

	"github.com/granate/schema"
)

func main() {
	schema.Provider.Query = Query{}
	fmt.Println("vim-go")
}
