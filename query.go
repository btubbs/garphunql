package garphunql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
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

func wrapQuery(query string) ([]byte, error) {
	return json.Marshal(map[string]string{"query": query})
}

// RawRequest takes a byte slice with your graphQL query inside it, and returns a byte slice with
// the graphql response inside it, or an error.
func (c *Client) RawRequest(query string) ([]byte, error) {
	q, err := wrapQuery(query)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(q)
	req, err := http.NewRequest("POST", c.url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
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
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("could not ready body: %v", err)
		}
		return nil, fmt.Errorf("non-200 response status: %v.  body: %v", resp.Status, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

// Request takes one Field object and the object or pointer that you want to have the results
// unmarshaled into.  It then does the request and unmarshals the result for you.
func (c *Client) Request(f Field, out interface{}) error {

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

// QueryFields takes one or more Req objects, each containing a field and an object or pointer that
// that field's data should be unmarshaled into.  It then joins all the fields into a single query,
// sends it to the server, and unmarshals the results into the containers you provided.
func (c *Client) QueryFields(first Q, more ...Q) error {
	reqs := map[string]Q{first.Field.Name: first}
	for _, f := range more {
		reqs[f.Field.Name] = f
	}

	// build an outer "query" with all the requested fields as sub selects
	fields := []Field{}
	for _, v := range reqs {
		fields = append(fields, v.Field)
	}
	q := Field{
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

// Q is a GraphQL Field object and a pointer to the thing you want to unmarshal its result into.
type Q struct {
	Field Field
	Dest  interface{}
}

// Field is a graphQL field.
type Field struct {
	Name      string
	Arguments map[string]interface{}
	Fields    []Field
}

// Render turns a Field into bytes that you can send in a network request.
func (f Field) Render(indents ...bool) (string, error) {
	out := f.Name
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

func wrapArgs(start string, things []string, sep string, end string) string {
	return start + strings.Join(things, sep) + end
}
