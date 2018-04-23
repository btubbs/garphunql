package gqlquery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	url     string
	headers map[string]string
}

func NewClient(url string, headers map[string]string) *Client {
	return &Client{
		url:     url,
		headers: headers,
	}
}

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

func (c *Client) Query(f *Field, out interface{}) error {
	q, err := f.Render()
	if err != nil {
		return err
	}

	// make request
	// scan result into provided output thing
	return nil
}

type Field struct {
	Name      string
	Arguments map[string]interface{}
	Fields    []Field
}

func (f Field) Render(indents ...bool) ([]byte, error) {
	out := []byte(f.Name)
	args := [][]byte{}
	for k, v := range f.Arguments {
		a := []byte(k + ": ")
		val, err := json.Marshal(v)
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
