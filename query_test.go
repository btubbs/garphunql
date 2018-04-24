package garphunql

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldRender(t *testing.T) {
	f := Field{
		Name: "query",
		Arguments: map[string]interface{}{
			"bar":  "baz",
			"quux": 1234,
		},
		Fields: []Field{
			{Name: "first"},
			{Name: "second"},
			{
				Name: "third",
				Arguments: map[string]interface{}{
					"arg1": "one",
					"arg2": 2,
				},
				Fields: []Field{
					{Name: "one"},
					{Name: "two"},
				},
			},
		},
	}

	rendered, err := f.Render()
	assert.Nil(t, err)
	expected := `query(bar: "baz", quux: 1234) {
  first
  second
  third(arg1: "one", arg2: 2) {
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
	resp, err := client.RawRequest([]byte(rawReq))
	assert.Nil(t, err)
	assert.Equal(t, fakeSuccessPayload, string(resp))
}

// test Request
func TestClientRequest(t *testing.T) {
	handler := http.HandlerFunc(fakeSuccessHandler)
	server := httptest.NewServer(&handler)
	defer server.Close()

	f := Field{
		Name: "query",
		Arguments: map[string]interface{}{
			"bar":  "baz",
			"quux": 1234,
		},
		Fields: []Field{
			{Name: "first"},
			{Name: "second"},
			{
				Name: "third",
				Arguments: map[string]interface{}{
					"arg1": "one",
					"arg2": 2,
				},
				Fields: []Field{
					{Name: "one"},
					{Name: "two"},
				},
			},
		},
	}

	var out fakeSuccessObject

	client := NewClient(server.URL, map[string]string{})
	err := client.Request(&f, &out)
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

// test MultiRequest
func TestClientMultiRequest(t *testing.T) {
	handler := http.HandlerFunc(fakeSuccessHandler)
	server := httptest.NewServer(&handler)
	defer server.Close()

	firstField := &Field{Name: "first"}
	secondField := &Field{Name: "second"}
	thirdField := &Field{
		Name: "third",
		Arguments: map[string]interface{}{
			"arg1": "one",
			"arg2": 2,
		},
		Fields: []Field{
			{Name: "one"},
			{Name: "two"},
		},
	}

	var first int
	var second string
	var thirdVal third

	client := NewClient(server.URL, map[string]string{})
	err := client.MultiRequest(
		&Req{Field: firstField, Dest: &first},
		&Req{Field: secondField, Dest: &second},
		&Req{Field: thirdField, Dest: &thirdVal},
	)
	assert.Nil(t, err)
	assert.Equal(t, 1, first)
	assert.Equal(t, "two", second)
	assert.Equal(t, third{One: 1, Two: 2}, thirdVal)
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
