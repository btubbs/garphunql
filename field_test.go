package garphunql

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldRender(t *testing.T) {
	f := Field("query",
		Arg("bar", "baz"),
		Arg("quux", 1234),
		GraphQLField{Name: "first", Alias: "foo"},
		"second",
		Field("third",
			Alias("somealias"),
			Arg("arg1", "one"),
			Arg("arg2", 2),
			"one",
			"two",
		),
	)

	rendered, err := f.Render()
	assert.Nil(t, err)
	expected := `query(bar: "baz", quux: 1234) {
  foo: first
  second
  somealias: third(arg1: "one", arg2: 2) {
    one
    two
  }
}`
	assert.Equal(t, expected, string(rendered))
}

// test RawRequest
func TestClientRawRequest(t *testing.T) {
	handler := http.HandlerFunc(fakeSuccessHandler)
	server := httptest.NewServer(&handler)
	defer server.Close()

	rawReq := `query(bar: "baz", quux: 1234) {
  first
  second
  third(arg1: "one", arg2: 2) {
    one
    two
  }
}`

	client := NewClient(server.URL, map[string]string{})
	resp, err := client.RawRequest(rawReq)
	assert.Nil(t, err)
	assert.Equal(t, fakeSuccessPayload, string(resp))
}

// test Request
func TestClientRequest(t *testing.T) {
	handler := http.HandlerFunc(fakeSuccessHandler)
	server := httptest.NewServer(&handler)
	defer server.Close()

	f := GraphQLField{
		Name: "query",
		Arguments: map[string]interface{}{
			"bar":  "baz",
			"quux": 1234,
		},
		Fields: []GraphQLField{
			{Name: "first"},
			{Name: "second"},
			{
				Name: "third",
				Arguments: map[string]interface{}{
					"arg1": "one",
					"arg2": 2,
				},
				Fields: []GraphQLField{
					{Name: "one"},
					{Name: "two"},
				},
			},
		},
	}

	var out fakeSuccessObject

	client := NewClient(server.URL, map[string]string{})
	err := client.Request(f, &out)
	assert.Nil(t, err)

	expected := fakeSuccessObject{
		Data: data{
			First:  1,
			Second: "two",
			Third: third{
				One: 1,
				Two: 2,
			},
		},
	}
	assert.Equal(t, expected, out)
}

// test Query
func TestClientQuery(t *testing.T) {
	handler := http.HandlerFunc(fakeSuccessHandler)
	server := httptest.NewServer(&handler)
	defer server.Close()

	firstField := GraphQLField{Name: "first"}
	secondField := GraphQLField{Name: "second"}
	thirdField := GraphQLField{
		Name: "third",
		Arguments: map[string]interface{}{
			"arg1": "one",
			"arg2": 2,
		},
		Fields: []GraphQLField{
			{Name: "one"},
			{Name: "two"},
		},
	}

	var first int
	var second string
	var thirdVal third

	client := NewClient(server.URL, map[string]string{})
	err := client.Query(
		Dest(firstField, &first),
		Dest(secondField, &second),
		Dest(thirdField, &thirdVal),
	)
	assert.Nil(t, err)
	assert.Equal(t, 1, first)
	assert.Equal(t, "two", second)
	assert.Equal(t, third{One: 1, Two: 2}, thirdVal)
}

func TestClientMutation(t *testing.T) {
	handler := http.HandlerFunc(fakeSuccessHandler)
	server := httptest.NewServer(&handler)
	defer server.Close()
	client := NewClient(server.URL, map[string]string{})
	firstField := GraphQLField{Name: "blah", Alias: "first"}
	var first int
	err := client.Mutation(
		Dest(firstField, &first),
	)
	assert.Equal(t, 1, first)
	assert.Nil(t, err)
}

// a handler that always returns a hardcoded payload.
const fakeSuccessPayload = `
			{
				"data": {
					"first": 1,
					"second": "two",
					"third": {
						"one": 1,
						"two": 2
					}
				}
			}
`

func fakeSuccessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/graphql")
	w.Write([]byte(fakeSuccessPayload))
}

type third struct {
	One int `json:"one"`
	Two int `json:"two"`
}

type data struct {
	First  int    `json:"first"`
	Second string `json:"second"`
	Third  third  `json:"third"`
}

type fakeSuccessObject struct {
	Data data `json:"data"`
}
