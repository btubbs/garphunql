package garphunql

type FieldOption interface {
	UpdateField(GraphQLField) GraphQLField
}

type Alias string

func (a Alias) UpdateField(f GraphQLField) GraphQLField {
	f.Alias = a
	return f
}

// An Argument represents a GraphQL argument, containing a name and value.
type Argument struct {
	Name  string
	Value interface{}
}

func (a Argument) UpdateField(f GraphQLField) GraphQLField {
	f.Arguments[a.Name] = a.Value
	return f
}

// Arg is a shorthand for building an Argument.
func Arg(name string, value interface{}) Argument {
	return Argument{
		Name:  name,
		Value: value,
	}
}

// The DestSetter type exists only so we can stick a UpdateField method on it.
type DestSetter func(GraphQLField) GraphQLField

// UpdateField sets the Dest field on a GraphQLField.
func (d DestSetter) UpdateField(f GraphQLField) GraphQLField {
	return d(f)
}

// Dest is a helper function for making a DestSetter.  Call it with a pointer to a thing as the
// argument, and use the return value as an argument to Query or Mutation.
func Dest(d interface{}) DestSetter {
	return func(f GraphQLField) GraphQLField {
		f.Dest = d
		return f
	}
}

// Field is a shorthand for building a GraphQLField.  It accepts a name, and any number of Arguments
// and GraphQLFields.  It returns a FieldFunc that can be Render-ed directly, or called with more
// Arguments and GraphQLFields to return a Render-able GraphQLField.
func Field(name string, outerOptions ...FieldOption) FieldFunc {
	return func(innerOptions ...FieldOption) GraphQLField {
		f := GraphQLField{
			Name:      name,
			Arguments: map[string]interface{}{},
			Fields:    []GraphQLField{},
		}
		for _, o := range outerOptions {
			f = o.UpdateField(f)
		}
		for _, o := range innerOptions {
			f = o.UpdateField(f)
		}
		return f
	}
}

// A FieldFunc can be Render-ed directly, or called with any number of Arguments and GraphQLFields
// to return a new Render-able GraphQLField.
type FieldFunc func(...FieldOption) GraphQLField

// Render takes a variable number of bools that indicate the number of indents to use in the query
// (their value doesn't matter), and returns the rendered field, or an error.
func (f FieldFunc) Render(indents ...bool) (string, error) {
	return f().Render(indents...)
}

// GetKey executes the FieldFunc and then returns the result of calling its GetKey method.
func (f FieldFunc) GetKey() string {
	return f().GetKey()
}

// Field executes the FieldFunc with no arguments and returns its resulting Field.
func (f FieldFunc) Field() GraphQLField {
	return f()
}

// UpdateField makes FieldFunc satisfy the FieldOption interface, so FieldFuncs can plug themselves
// into parents as sub selections.
func (f FieldFunc) UpdateField(parent GraphQLField) GraphQLField {
	return f().UpdateField(parent)
}
