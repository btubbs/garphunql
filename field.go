package garphunql

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// GraphQLField is a graphQL field.  Normally you will have these built for you by passing arguments
// to FieldFuncs instead of constructing them directly.
type GraphQLField struct {
	Name      string
	Arguments map[string]interface{}
	Fields    []GraphQLField
	Alias     Alias
}

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
		val, err := json.Marshal(f.Arguments[k])
		if err != nil {
			return "", err
		}
		a += string(val)
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

func (f GraphQLField) GetKey() string {
	if len(f.Alias) > 0 {
		return string(f.Alias)
	}
	return f.Name
}

func wrapArgs(start string, things []string, sep string, end string) string {
	return start + strings.Join(things, sep) + end
}
