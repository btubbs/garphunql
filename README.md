# Garphunql [![Build Status](https://travis-ci.org/btubbs/garphunql.svg?branch=master)](https://travis-ci.org/btubbs/garphunql) [![Coverage Status](https://coveralls.io/repos/github/btubbs/garphunql/badge.svg?branch=master)](https://coveralls.io/github/btubbs/garphunql?branch=master)

Garphunql is a Golang client library for GraphQL.  It supports building queries in a type-safe way,
then submitting them to the server and automatically unmarshaling the responses.

## Usage

### Getting a Client

All of Garphunql's functionality is provided as methods on a `garphunql.Client` object.  You
instantiate a client with the `garphunql.NewClient` function, providing the server URL and a map of
headers:

```go
client := garphunql.NewClient(
  "https://api.github.com/graphql",
  map[string]string{
    "Authorization": "bearer aidee6gahPe1baeth8tikeijeeth0aedaehe",
  },
)
```

### client.QueryFields

Garphunql supports making queries in several ways.  The highest level and most convenient way is
the `QueryFields` method, which bundles queries on several fields together, then splits out the
results and unmarshals them into pointers that you have provided.  When calling `QueryFields`, you
provide one or more pairs of `Field` objects and the things they should be scanned into.  Here's an
example that gets information on a couple users and all the open source licenses on Github.com:

```go
package main

import (
  "fmt"

  "github.com/btubbs/garphunql"
)

func main() {

  // First define a few fields that we want to query for.  This first one just tells about the
  // currently logged in user.
  meField := garphunql.Field{
    Name: "viewer",
    Fields: []garphunql.Field{
      {Name: "name"},
      {Name: "location"},
    },
  }

  // This field is a query for another user.  Github requires that we pass a 'login' argument to
  // identify the user.
  zachField := garphunql.Field{
    Name: "user",
    Arguments: map[string]interface{}{
      "login": "zachabrahams",
    },
    Fields: []garphunql.Field{
      {Name: "name"},
      {Name: "location"},
    },
  }

  // Also, let's get the names and permissions of all the open source licenses that Github knows
  // about.  This demonstrates querying more deeply nested fields, and getting a list of results
  // back instead of a single object.
  licensesField := garphunql.Field{
    Name: "licenses",
    Fields: []garphunql.Field{
      {Name: "name"},
      {
        Name: "permissions",
        Fields: []garphunql.Field{
          {Name: "description"},
        },
      },
    },
  }

  client := garphunql.NewClient(
    "https://api.github.com/graphql",
    map[string]string{
      "Authorization": "bearer aidee6gahPe1baeth8tikeijeeth0aedaehe",
    },
  )

  var me User
  var zach User
  var licenses []License

  // This right here is the cool bit.  The client.QueryFields call will combine all the fields you
  // provide into one query, send it to the server, and then split up the results into the pointers
  // that you provided with each field.
  err := client.QueryFields(
    garphunql.Q{Field: meField, Dest: &me},
    garphunql.Q{Field: zachField, Dest: &zach},
    garphunql.Q{Field: licensesField, Dest: &licenses},
  )

  if err != nil {
    panic(err.Error())
  }

  // our structs and slice are now populated!
  fmt.Println("me", me)
  fmt.Println("zach", zach)
  fmt.Println("licenses", licenses)
}

type License struct {
  Name        string       `json:"name"`
  Permissions []Permission `json:"permissions"`
}

type Permission struct {
  Description string `json:"description"`
}

type User struct {
  Name     string `json:"name"`
  Location string `json:"location"`
}
```    

### client.Request

If you only need to query a single field, or if you want to run a mutation, then the `Request`
method can do it.  It takes only a single `Field`, and something to unmarshal the results into.
Here's an example of getting a single user from the Github API using `Request`:

```go
package main

import (
  "fmt"

  "github.com/btubbs/garphunql"
)

func main() {
  client := garphunql.NewClient(
    "https://api.github.com/graphql",
    map[string]string{
      "Authorization": "bearer aidee6gahPe1baeth8tikeijeeth0aedaehe",
    },
  )
  query := garphunql.Field{
    Name: "query",
    Fields: []garphunql.Field{
      {
        Name: "user",
        Arguments: map[string]interface{}{
          "login": "zachabrahams",
        },
        Fields: []garphunql.Field{
          {Name: "name"},
          {Name: "location"},
        },
      },
    },
  }

  var resp Resp
  err := client.Request(query, &resp)
  fmt.Println(resp, err)
}

type Resp struct {
  Data struct {
    User User `json:"user"`
  } `json:"data"`
}

type User struct {
  Name     string `json:"name"`
  Location string `json:"location"`
}
```

Notice that `Request`, unlike `QueryFields`, requires that your top level field be named `query`,
just like when making raw GraphQL requests.  It also requires that you handle unwrapping the outer
level `data` object returned by the GraphQL server (which is done in the example above with the
`Resp` struct).

### client.RawRequest

The lowest level interface offered by Garphunql is the `RawRequest` method, which takes a query as a
string and returns the exact bytes returned by the server:

```go
package main

import (
  "fmt"

  "github.com/btubbs/garphunql"
)

func main() {
  client := garphunql.NewClient(
    "https://api.github.com/graphql",
    map[string]string{
      "Authorization": "bearer aidee6gahPe1baeth8tikeijeeth0aedaehe",
    },
  )

  q := `{
user(login:"zachabrahams") {
  name
  location
}
}`

  resp, err := client.RawRequest(q)
  fmt.Println(string(resp), err)
}
```

## TODO
- request variables
- input objects
- `... on` syntax
