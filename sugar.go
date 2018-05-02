package garphunql

import (
	"fmt"
	"reflect"
)

type Alias string

// An Argument represents a GraphQL argument, containing a name and value.
type Argument struct {
	Name  string
	Value interface{}
}

// Arg is a shorthand for building an Argument.
func Arg(name string, value interface{}) Argument {
	return Argument{
		Name:  name,
		Value: value,
	}
}

// Field is a shorthand for building a GraphQLField.  It accepts a name, and any number of Arguments
// and GraphQLFields.  It returns a FieldFunc that can be Render-ed directly, or called with more
// Arguments and GraphQLFields to return a Render-able GraphQLField.
func Field(name string, outerFieldsAndArgs ...interface{}) FieldFunc {
	return func(innerFieldsAndArgs ...interface{}) GraphQLField {
		f := GraphQLField{
			Name:      name,
			Arguments: map[string]interface{}{},
			Fields:    []GraphQLField{},
		}
		appendArgsFields(&f, outerFieldsAndArgs)
		appendArgsFields(&f, innerFieldsAndArgs)
		return f
	}
}

func appendArgsFields(f *GraphQLField, fieldsAndArgs []interface{}) {
	for _, thing := range fieldsAndArgs {
		switch v := thing.(type) {
		case GraphQLField:
			f.Fields = append(f.Fields, v)
		case FieldFunc:
			f.Fields = append(f.Fields, v())
		case string:
			f.Fields = append(f.Fields, GraphQLField{Name: v})
		case Argument:
			f.Arguments[v.Name] = v.Value
		case Alias:
			f.Alias = v
		default:
			panic(fmt.Sprintf("cannot handle value %v, type %v", v, reflect.TypeOf(v)))
		}
	}
}

// A FieldFunc can be Render-ed directly, or called with any number of Arguments and GraphQLFields
// to return a new Render-able GraphQLField.
type FieldFunc func(...interface{}) GraphQLField

// Render takes a variable number of bools that indicate the number of indents to use in the query
// (their value doesn't matter), and returns the rendered field, or an error.
func (f FieldFunc) Render(indents ...bool) (string, error) {
	return f().Render(indents...)
}

func (f FieldFunc) GetKey() string {
	return f().GetKey()
}

func Dest(f Fielder, d interface{}) FieldDest {
	return FieldDest{Field: f, Dest: d}
}
