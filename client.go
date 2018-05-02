package garphunql

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

type Fielder interface {
	Render(...bool) (string, error)
	GetKey() string
}

// Request takes one GraphQLField object and the object or pointer that you want to have the results
// unmarshaled into.  It then does the request and unmarshals the result for you.
func (c *Client) Request(f Fielder, out interface{}) error {

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

// Query takes one or more FieldDest objects, each containing a field and an object or pointer that
// that field's data should be unmarshaled into.  It then joins all the fields into a single query,
// sends it to the server, and unmarshals the results into the containers you provided.
func (c *Client) Query(first FieldDest, more ...FieldDest) error {
	reqs := map[string]FieldDest{first.Field.GetKey(): first}
	for _, f := range more {
		reqs[f.Field.GetKey()] = f
	}

	// build an outer "query" with all the requested fields as sub selects
	fields := []GraphQLField{}
	for _, v := range reqs {
		switch f := v.Field.(type) {
		case GraphQLField:
			fields = append(fields, f)
		case FieldFunc:
			fields = append(fields, f())
		}
	}
	q := GraphQLField{
		Name:   "query",
		Fields: fields,
	}

	return c.queryFields(q, reqs)
}

func (c *Client) queryFields(q GraphQLField, reqs map[string]FieldDest) error {

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

func (c *Client) Mutation(f FieldDest) error {
	var field GraphQLField
	switch v := f.Field.(type) {
	case GraphQLField:
		field = v
	case FieldFunc:
		field = v()
	}
	q := GraphQLField{
		Name:   "mutation",
		Fields: []GraphQLField{field},
	}

	return c.queryFields(q, map[string]FieldDest{f.Field.GetKey(): f})
}

type genericResult struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []gqlError                 `json:"errors"`
}

type gqlError struct {
	Message string `json:"message"`
}

// FieldDest is a GraphQLField object and a pointer to the thing you want to unmarshal its result into.
type FieldDest struct {
	Field Fielder
	Dest  interface{}
}
