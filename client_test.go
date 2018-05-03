package garphunql

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://foo.com", Header("foo", "bar"))
	assert.Equal(t, client.url, "http://foo.com")
	assert.Equal(t, client.headers, map[string]string{"foo": "bar"})
}

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

	client := NewClient(server.URL)
	resp, err := client.RawRequest(rawReq)
	assert.Nil(t, err)
	assert.Equal(t, fakeSuccessPayload, string(resp))
}

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

	client := NewClient(server.URL)
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

func TestClientQuery(t *testing.T) {
	handler := http.HandlerFunc(fakeSuccessHandler)
	server := httptest.NewServer(&handler)
	defer server.Close()

	var first int
	var second string
	var thirdVal third

	firstField := GraphQLField{Name: "first", Dest: &first}
	secondField := GraphQLField{Name: "second", Dest: &second}
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
		Dest: &thirdVal,
	}

	client := NewClient(server.URL)
	err := client.Query(
		firstField,
		secondField,
		thirdField,
	)
	assert.Nil(t, err)
	assert.Equal(t, 1, first)
	assert.Equal(t, "two", second)
	assert.Equal(t, third{One: 1, Two: 2}, thirdVal)
}

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
	w.Header().Set("Content-Type", "application/json")
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

func TestClientMutation(t *testing.T) {
	handler := http.HandlerFunc(fakeSuccessHandler)
	server := httptest.NewServer(&handler)
	defer server.Close()
	client := NewClient(server.URL)
	var first int
	firstField := GraphQLField{Name: "blah", Alias: "first", Dest: &first}
	err := client.Mutation(
		firstField,
	)
	assert.Equal(t, 1, first)
	assert.Nil(t, err)
}
