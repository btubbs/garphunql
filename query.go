package gqlquery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Client is the object used for making all requests.
type Client struct {
	url     string
	headers map[string]string
}

// NewClient returns a new client object.
func NewClient(url string, headers map[string]string) *Client {
	return &Client{
		url:     url,
		headers: headers,
	}
}

// RawRequest takes a byte slice with your graphQL query inside it, and returns a byte slice with
// the graphql response inside it, or an error.
func (c *Client) RawRequest(query []byte) ([]byte, error) {
	buf := bytes.NewBuffer(query)
	req, err := http.NewRequest("POST", c.url, buf)
	req.Header.Add("Content-Type", "application/graphql")
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response status: %v", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

// Request takes one Field object and the object or pointer that you want to have the results
// unmarshaled into.  It then does the request and unmarshals the result for you.
func (c *Client) Request(f *Field, out interface{}) error {

	query, err := f.Render()
	if err != nil {
		return err
	}

	// make request
	res, err := c.RawRequest(query)
	if err != nil {
		return err
	}
	// scan result into provided output thing
	return json.Unmarshal(res, out)
}

// MultiRequest takes one or more Req objects, each containing a field and an object or pointer that
// that field's data should be unmarshaled into.  It then joins all the fields into a single query,
// sends it to the server, and unmarshals the results into the containers you provided.
func (c *Client) MultiRequest(first *Req, more ...*Req) error {
	reqs := map[string]*Req{first.Field.Name: first}
	for _, f := range more {
		reqs[f.Field.Name] = f
	}

	// build an outer "query" with all the requested fields as sub selects
	fields := []Field{}
	for _, v := range reqs {
		fields = append(fields, *v.Field)
	}
	q := &Field{
		Name:   "query",
		Fields: fields,
	}

	res := genericResult{}
	err := c.Request(q, &res)
	if err != nil {
		return err
	}

	// now loop over given requests and pluck/unmarshall the payloads for each one
	for k, v := range reqs {
		err := json.Unmarshal(res.Data[k], v.Dest)
		if err != nil {
			return err
		}
	}
	return nil
}

type genericResult struct {
	Data map[string]json.RawMessage `json:"data"`
}

// Req is a GraphQL Field object and a pointer to the thing you want to unmarshal its result into.
type Req struct {
	Field *Field
	Dest  interface{}
}

// Field is a graphQL field.
type Field struct {
	Name      string
	Arguments map[string]interface{}
	Fields    []Field
}

// Render turns a Field into bytes that you can send in a network request.
func (f Field) Render(indents ...bool) ([]byte, error) {
	out := []byte(f.Name)
	// loop over the args in alphabetical order to ensure consistent (easily testable) output order.
	argNames := []string{}
	for k, _ := range f.Arguments {
		argNames = append(argNames, k)
	}
	args := [][]byte{}
	for _, k := range argNames {
		a := []byte(k + ": ")
		val, err := json.Marshal(f.Arguments[k])
		if err != nil {
			return nil, err
		}
		a = append(a, val...)
		args = append(args, a)
	}

	// only include args part if there's at least one arg
	if len(args) > 0 {
		out = append(out, wrapArgs("(", args, ", ", ")")...)
	}

	// ok now render sub selects.
	subfields := [][]byte{}
	for _, s := range f.Fields {
		val, err := s.Render(append(indents, true)...)
		if err != nil {
			return nil, err
		}
		subfields = append(subfields, val)
	}
	if len(subfields) > 0 {
		indent := []byte("  ")
		// curly brace
		out = append(out, []byte(" {")...)
		// first indent
		// now render each field
		for _, f := range subfields {
			out = append(out, []byte("\n")...)
			out = append(out, bytes.Repeat(indent, len(indents)+1)...)
			out = append(out, f...)
		}
		out = append(out, []byte("\n")...)
		out = append(out, bytes.Repeat(indent, len(indents))...)
		out = append(out, []byte("}")...)
	}
	return out, nil
}

func wrapArgs(start string, things [][]byte, sep string, end string) []byte {
	b := []byte(start)
	b = append(b, bytes.Join(things, []byte(sep))...)
	return append(b, []byte(end)...)
}
