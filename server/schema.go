package main

import "github.com/graphql-go-gen/schema"

var _ schema.UserInterface = (*User)(nil)

type User struct {
	ID    string
	Name  string
	Todos []schema.TodoInterface
}

func (user User) NameField() (string, error) {
	return user.Name, nil
}

func (user User) IdField() (string, error) {
	return user.ID, nil
}

func (user User) TodosField() ([]schema.TodoInterface, error) {
	return user.Todos, nil
}

var _ schema.TodoInterface = (*Todo)(nil)

type Todo struct {
	ID          string
	Title       string
	Description string
	Status      int
}

func (todo Todo) IdField() (string, error) {
	return todo.ID, nil
}

func (todo Todo) TitleField() (string, error) {
	return todo.Title, nil
}

func (todo Todo) DescriptionField() (string, error) {
	return todo.Description, nil
}

func (todo Todo) StatusField() (int, error) {
	return todo.Status, nil
}

var _ schema.QueryInterface = (*Query)(nil)

type Query struct {
	User User
}

func (query Query) ViewerField() (schema.UserInterface, error) {
	return query.User, nil
}
