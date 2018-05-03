# Garphunql [![Build Status](https://travis-ci.org/btubbs/garphunql.svg?branch=master)](https://travis-ci.org/btubbs/garphunql) [![Coverage Status](https://coveralls.io/repos/github/btubbs/garphunql/badge.svg?branch=master)](https://coveralls.io/github/btubbs/garphunql?branch=master)

Garphunql is a Golang client library for GraphQL.  It supports building queries in a type-safe way,
then submitting them to the server and automatically unmarshaling the responses.

## Usage

### Getting a Client

Garphunql's functionality is provided as methods on a `Client` object.  You instantiate a client
with the `NewClient` function, providing the server URL and any number of headers.  Here's an
example of making a client to talk to the Github GraphQL API.  (In all the examples here, garphunql
has been imported with the `gql` alias.):

```go
package main

import (
  "fmt"

  gql "github.com/btubbs/garphunql"
)

func main() {
	client := gql.NewClient(
		"https://api.github.com/graphql",
		gql.Header("Authorization", "bearer aidee6gahPe1baeth8tikeijeeth0aedaehe"),
	)
  // ...
}
```

### Making a simple query

Once you have a client you can query for data using the client's `Query` method.  The Query method
takes a field, which you can construct with the `Field` function, passing it a name, the sub-fields
you want to get back, and a destination that the results should be unmarshaled into:

```go
type User struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

func simpleQuery(client *gql.Client) {
	var me User
	meField := gql.Field("viewer",
		gql.Field("name"),
		gql.Field("location"),
		gql.Dest(&me),
	)
	err := client.Query(meField)
	fmt.Println(err, me)
}
```

### Passing arguments to fields

Some GraphQL fields take arguments.  Garphunql supports querying those fields by passing in one or
more `Arg` calls to `Field`.  Here we query Github for another user, passing in a value for the
`login` argument:

```go
func queryWithArguments(client *gql.Client) {
	var zach User
	zachField := gql.Field("user",
		gql.Arg("login", "zachabrahams"),
		gql.Field("name"),
		gql.Field("location"),
		gql.Dest(&zach),
	)
	err := client.Query(zachField)
	fmt.Println(err, zach)
}
```

### More deeply nested fields

GraphQL can have fields within fields within fields, etc.  This example shows a `label` field nested
inside a `permissions` field nested inside a `licenses` field (which is also automatically nested
inside a `query` field before Garphunql sends it over the wire):

```go
func deeplyNestedFields(client *gql.Client) {
	var licenses []License
	licensesField := gql.Field("licenses",
		gql.Field("name"),
		gql.Field("permissions",
      gql.Field("label"),
    ),
	)
	err := client.Query(
		licensesField(gql.Dest(&licenses)),
	)
	fmt.Println(err, licenses)
}
```

### Querying multiple fields at once

GraphQL lets you query any number of fields at the same time.  Similarly, Garphunql lets you pass in
any number of `Field` calls to `Query`.  The fields will all be bundled together and sent to the
server as sub-fields of the top-level "query" field.  When the server's response is received, each
piece of the payload will be unmarshaled into the appropriate destination:

```go
func multipleQueries(client *gql.Client) {
	var me User
	var zach User
	meField := gql.Field("viewer",
		gql.Field("name"),
		gql.Field("location"),
		gql.Dest(&me),
	)
	zachField := gql.Field("user",
		gql.Arg("login", "zachabrahams"),
		gql.Field("name"),
		gql.Field("location"),
		gql.Dest(&zach),
	)
	err := client.Query(meField, zachField)
	fmt.Println(err, me, zach)
}
```

### Re-using fields with late binding

You might want to query the same field multiple times with different arguments.  Garphunql lets you
partially define a field with the options shared between your calls, then call it later with more
arguments to customize it:
```go
func lateBoundFields(client *gql.Client) {
	unboundUserField := gql.Field("user",
		gql.Field("name"),
		gql.Field("location"),
	)
	var dave User
	var kelsey User
	err := client.Query(
		unboundUserField(gql.Arg("login", "davecheney"), gql.Dest(&dave)),
		unboundUserField(gql.Arg("login", "kelseyhightower"), gql.Dest(&kelsey)),
	)
	fmt.Println(err, dave, kelsey)
}

```

There's one more thing to note about that example.  We made two queries to the `user` field.
Normally GraphQL would require you to provide an alias for at least one of them so that they could
be differentiated in the response payload.  Garphunql automatically detected the name collision
while building the query and set an alias for the second `user` field behind the scenes.

### Handling resolver errors

In a GraphQL server, every field is populated by a "resolver".  Because you can query multiple
fields at once, it's possible for some of their resolvers to succeed, while others fail.  In this
case, the server will return `null` for the failed fields, and add objects to the "errors" element
of the response payload.  Garphunql bundles those errors into a
[multierror](https://github.com/hashicorp/go-multierror) and returns it as the result of the `Query`
call, while still populating the destination variables of the fields that succeeded:

```go
func errorHandling(client *gql.Client) {
	var me User
	var nobody User
	meField := gql.Field("viewer",
		gql.Field("name"),
		gql.Field("location"),
		gql.Dest(&me),
	)
	userField := gql.Field("user",
		gql.Arg("login", "TOTALLY FAKE USER"),
		gql.Field("name"),
		gql.Field("location"),
		gql.Dest(&nobody),
	)
	err := client.Query(
		meField,
		userField,
	)
	fmt.Println(err, me, nobody)
}

```

The function above will print the error returned by Github (`* Could not resolve to a User with the
login of 'TOTALLY FAKE USER'.`), as well as printing the populated `me` and the empty `nobody`.

### Mutations

In addition to the `Query` method, the Garphunql client provides a `Mutation` method.  It works
identically, except that you may only provide a single top-level field.

## TODO
- request variables
- input objects
- `... on` syntax
