# !WORK IN PROGRESS!

# Granate - Code generator for graphql
`Granate` is a tool meant to speed up development of applications that utilises
graphql. `Granate` takes a graphql `schema` and outputs code based on the
specified language.

### Supported languages

- Go

## Quick start
### Install
```sh
# A dependency
go get golang.org/x/tools/cmd/goimports

# The granate package
go get github.com/granateio/granate
```

### Usage
`Granate` requires a config yaml file `granate.yaml` to provide some basic
information about the project.
```yaml
# Select language to output
language: go

# Output locations (and some options)
output:
  # Output target directory (value can be empty,
  # but the option is currently required)
  target: <target-dir>

  # Name of the package
  package: change.me/your/package-name/<target-dir>

  # Name of the package to generate for the schema
  schema: schema

  # Name of the package to generate for the models
  models: models

# Schemas to use for the code generator
schemas:
  - schema.graphql
```

A schema is also required, you can provide multiple schemas in the `schemas`
section of the config file. Here is a simple `todo.graphql` file
```graphql
# A user in the system
type User {
    id: ID
    name: String
    todos: [Todo]
}

# A todo
type Todo {
    id: ID
    title: String
    description: String
}

# Root query
type Query {
    # Get the current loged in user
    viewer: User
}

```

By simply running `granate` in the same folder as the `granate.yaml` or placing
`//go:generate granate` at the top of your `main.go` file and running `go
generate`, three files will be created.
```
schema/
- definitions.go
- adapters.go
- provider.go
```

The `definitions.go` file is where all the graphql specific code is.
`adapters.go` provides a set of interfaces to use for implementing the logic.
`provider.go` contains a set of function to bootstrap the graphql schema as
well as providing a graphiql interface to test your schema with.

For a more in depth overview of how to use `Granate`, check out the simple example under the `example` folder.

## Unsupported features

We support most graphql features. These are examples of things you can't yet
generate with granate.

```graphql
# Define interfaces (the only interface you can use for now is Node, which exists)
interface Entity {
    id: ID!
    name: String
}

# Define scalars
scalar URL

# Define unions
union Animal = Dog | Snake

# Define schemas
# You shouldn't need to do this
schema {
    query: Query
    mutation: Mutation
    subscription: ...
}
```

See the [GraphQL
cheat sheet](https://wehavefaces.net/graphql-shorthand-notation-cheatsheet-17cd715861b6)
for reference.
