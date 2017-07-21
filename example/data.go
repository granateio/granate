package main

import "github.com/granateio/granate/example/schema"

var users = []schema.UserInterface{
	User{
		Todos: todos,
		ID:    "1",
		Name:  "Jonas",
	},
	User{
		Todos: todos,
		ID:    "2",
		Name:  "Nicolai",
	},
	User{
		Todos: todos,
		ID:    "3",
		Name:  "Christian",
	},
}

var todos = []schema.TodoInterface{
	Todo{
		ID:          "1",
		Description: "A todo, pleas do",
		Status:      schema.ACTIVE,
		Title:       "Do todos",
	},
	Todo{
		ID:          "2",
		Description: "Another todo, pleas don't do",
		Status:      schema.PAUSED,
		Title:       "Don't do todos",
	},
}
