//go:generate granate

package main

import (
	users_models "github.com/granateio/granate/small/users/models"
	users_schema "github.com/granateio/granate/small/users/schema"
)

func main() {
	// up := userprovider.UserProvider{}
	// up.AddUser(userprovider.User{
	// 	ID:    "1",
	// 	Name:  "Jonas",
	// 	Email: "joans.rudlang@gmail.com",
	// })
	// up.AddUser(userprovider.User{
	// 	ID:    "2",
	// 	Name:  "Mem",
	// 	Email: "mem@mem.mem",
	// })
	// up.AddUser(userprovider.User{
	// 	ID:    "3",
	// 	Name:  "Nicolai",
	// 	Email: "nico@lai.skog",
	// })
	//
	// providers := Providers{
	// 	Users: up,
	// }

	// root := models.Root{
	// 	UserProvider: &providers.Users,
	// }
	root := users_models.Root{}

	users_schema.Init(users_schema.ProviderConfig{
		Query:    root,
		Mutation: root,
		Relay:    root,
	})

	users_schema.Serve(":8080")
}
