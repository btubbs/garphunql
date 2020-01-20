package garphunql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	multierror "github.com/hashicorp/go-multierror"
)

var seededRand *rand.Rand

const keyChars = "abcdefghijklmnopqrstuvwxyz"

// Client is the object used for making all requests.
type Client struct {
	url        string
	headers    map[string]string
	httpClient *http.Client
}

// A ClientOption is a function that modifies a *Client.
type ClientOption func(*Client)

// NewClient returns a new client object.
func NewClient(url string, options ...ClientOption) *Client {
	c := &Client{
		url:        url,
		headers:    map[string]string{},
		httpClient: &http.Client{},
	}
	for _, o := range options {
		o(c)
	}
	return c
}

func Header(key, val string) ClientOption {
	return func(c *Client) {
		c.headers[key] = val
	}
}

func HttpClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func wrapQuery(query string) ([]byte, error) {
	return json.Marshal(map[string]string{"query": query})
}

// RawRequest takes a byte slice with your graphQL query inside it, and returns a byte slice with
// the graphql response inside it, or an error.
func (c *Client) RawRequest(ctx context.Context, query string) ([]byte, error) {
	q, err := wrapQuery(query)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(q)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}

	resp, err := c.httpClient.Do(req)
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

	return ioutil.ReadAll(resp.Body)
}

// Fielder defines the functions that a thing must implement in order to be passed to our Query and
// Mutation functions.
type Fielder interface {
	Render(...bool) (string, error)
	GetKey() string
	Field() GraphQLField
}

// Request takes one GraphQLField object and the object or pointer that you want to have the results
// unmarshaled into.  It then does the request and unmarshals the result for you.
func (c *Client) Request(ctx context.Context, f Fielder, out interface{}) error {

	query, err := f.Render()
	if err != nil {
		return err
	}

	// make request
	res, err := c.RawRequest(ctx, query)
	if err != nil {
		return err
	}
	// scan result into provided output thing
	return json.Unmarshal(res, out)
}

// Query takes one or more GraphQLField objects.  It joins all the fields into a single query,
// sends it to the server, and unmarshals the results into Dest fields of the GraphQLField objects.
func (c *Client) Query(first Fielder, more ...Fielder) error {
	return c.QueryContext(context.Background(), first, more...)
}

func (c *Client) QueryContext(ctx context.Context, first Fielder, more ...Fielder) error {
	q, destMap := wrapFields("query", first, more...)
	return c.queryFields(ctx, q, destMap)
}

// Mutation accepts a GraphQLField, wraps it in a "mutation" field, performs the query, then scans
// the result into the field's dest.
func (c *Client) Mutation(f Fielder) error {
	return c.MutationContext(context.Background(), f)
}

func (c *Client) MutationContext(ctx context.Context, f Fielder) error {
	q, destMap := wrapFields("mutation", f)
	return c.queryFields(ctx, q, destMap)
}

func (c *Client) queryFields(ctx context.Context, q GraphQLField, destMap map[string]interface{}) error {
	res := GenericResult{}
	err := c.Request(ctx, q, &res)
	if err != nil {
		return err
	}

	// If there were any errors server side, then those fields will have come back as null, and they
	// should have entries in the response's "errors" list.  Combine and report any such server
	// errors.
	var errs *multierror.Error
	for _, e := range res.Errors {
		errs = multierror.Append(errs, e)
	}

	// now loop over given requests and pluck/unmarshall the payloads for each one
	for k, v := range destMap {
		if data, ok := res.Data[k]; ok {
			err := json.Unmarshal(data, v)
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("could not unmarshal %v (%s) into %v: %v", res.Data[k], k, v, err))
			}
		}
	}
	return errs.ErrorOrNil()
}

// GenericResult matches the outermost structure of a GraphQL response payload.
type GenericResult struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []GraphQLError             `json:"errors"`
}

// return a random string 8 chars long, for use as an alias in a query with a repeated field.
func randomKey() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = keyChars[seededRand.Intn(len(keyChars))]
	}
	return string(b)
}

func init() {
	// provide a pseudo random seed for generating field aliases, if necessary.
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}
