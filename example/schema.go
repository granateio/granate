package main

import (
	"fmt"
	"strconv"

	"github.com/granate/example/schema"
)

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
	Userid      int
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

func (todo Todo) UseridField() (int, error) {
	return todo.Userid, nil
}

var _ schema.QueryInterface = (*Query)(nil)

type Query struct {
	User schema.UserInterface
}

func (query Query) ViewerField() (schema.UserInterface, error) {
	return query.User, nil
}

var _ schema.MutationInterface = (*Mutation)(nil)

type Mutation struct {
}

func (mut Mutation) CreateTodosField(id string,
	todos []schema.TodoInputStruct) ([]schema.TodoInterface, error) {

	newTodos := make([]schema.TodoInterface, 0, len(todos))

	for i, todo := range todos {
		newTodos = append(newTodos, Todo{
			ID:          strconv.Itoa(i + 1),
			Description: todo.Description,
			Title:       todo.Title,
		})
	}

	return newTodos, nil
}

func (mut Mutation) ChangeTodoStatusField(id string, status int) (
	schema.TodoInterface, error) {
	fmt.Printf("User: %s, Status: %d\n", id, status)
	return Todo{Title: "Dummy"}, nil
}
