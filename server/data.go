package main

import "github.com/granate/schema"

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
		Name:  "Christion",
	},
}

var todos = []schema.TodoInterface{
	Todo{
		ID:          "1",
		Description: "A todo, pleas do",
		Status:      1,
		Title:       "Do todos",
	},
	Todo{
		ID:          "2",
		Description: "Another todo, pleas don't do",
		Status:      1,
		Title:       "Don't do todos",
	},
}
