package garphunql

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//ArgumentFormatter - interface for a format method
type ArgumentFormatter interface {
	Format() (string, error)
}

//Format - type for Enum
func (e Enum) Format() (string, error) {
	return string(e), nil
}

//Format - type for JSONMap
func (e JSONMap) Format() (string, error) {
	output := "{"
	items := []string{}
	for k, v := range e {
		i := k + ":"
		switch val := v.(type) {
		case ArgumentFormatter:
			result, err := val.Format()
			if err != nil {
				return "", err
			}
			i += result
		default:
			slice, err := json.Marshal(val)
			if err != nil {
				return "", err
			}
			i += string(slice)
		}
		items = append(items, i)
	}

	output += strings.Join(items, ",")
	output += "}"
	return output, nil
}

// GraphQLField is a graphQL field.  Normally you will have these built for you by passing arguments
// to FieldFuncs instead of constructing them directly.
type GraphQLField struct {
	Name      string
	Arguments map[string]interface{}
	Fields    []GraphQLField
	Alias     Alias
	Dest      interface{}
}

// Enum is a string type alias for argument values that shouldn't have quotes put around them.
type Enum string

// JSONMap correctly formats objects that are json types
type JSONMap map[string]interface{}

// Render turns a Field into bytes that you can send in a network request.
func (f GraphQLField) Render(indents ...bool) (string, error) {
	out := ""
	if len(f.Alias) > 0 {
		out += fmt.Sprintf("%s: ", f.Alias)
	}

	out += f.Name
	// loop over the args in alphabetical order to ensure consistent (easily testable) output order.
	argNames := []string{}
	for k, _ := range f.Arguments {
		argNames = append(argNames, k)
	}
	sort.Strings(argNames)
	args := []string{}
	for _, k := range argNames {
		a := k + ": "
		switch argVal := f.Arguments[k].(type) {
		case Enum:
			a += string(argVal)
		default:
			val, err := json.Marshal(argVal)
			if err != nil {
				return "", err
			}
			a += string(val)
		}
		args = append(args, a)
	}

	// only include args part if there's at least one arg
	if len(args) > 0 {
		out += wrapArgs("(", args, ", ", ")")
	}

	// ok now render sub selects.
	subfields := []string{}
	for _, s := range f.Fields {
		val, err := s.Render(append(indents, true)...)
		if err != nil {
			return "", err
		}
		subfields = append(subfields, val)
	}
	if len(subfields) > 0 {
		indent := "  "
		// curly brace
		out += " {"
		// first indent
		// now render each field
		for _, f := range subfields {
			out += "\n"
			out += strings.Repeat(indent, len(indents)+1)
			out += f
		}
		out += "\n"
		out += strings.Repeat(indent, len(indents))
		out += "}"
	}
	return out, nil
}

// GetKey returns the field's alias, if it has one, or otherwise its name.
func (f GraphQLField) GetKey() string {
	if len(f.Alias) > 0 {
		return string(f.Alias)
	}
	return f.Name
}

// Field returns the field's field.
func (f GraphQLField) Field() GraphQLField {
	return f
}

// UpdateField makes GraphQLField satisfy the FieldOption interface, which lets it plug itself into
// a parent field as a sub selection.
func (f GraphQLField) UpdateField(parent GraphQLField) GraphQLField {
	parent.Fields = append(parent.Fields, f)
	return parent
}

func wrapArgs(start string, things []string, sep string, end string) string {
	return start + strings.Join(things, sep) + end
}

// wrapFields takes a field name and one or more FieldDest pairs, and wraps them into a single
// GraphQLField and a map of the destinations for unmarshaling the results.
func wrapFields(name string, first Fielder, more ...Fielder) (GraphQLField, map[string]interface{}) {
	fields := []GraphQLField{first.Field()}
	dests := map[string]interface{}{first.GetKey(): first.Field().Dest}
	for _, f := range more {
		field := f.Field()
		key := f.GetKey()
		// if this field is already present in the dest map, automatically use an alias to
		// disambiguate this one.
		if _, ok := dests[key]; ok {
			key = randomKey()
			field.Alias = Alias(key)
		}
		dests[key] = field.Dest
		fields = append(fields, field)
	}

	return GraphQLField{Name: name, Fields: fields}, dests
}
