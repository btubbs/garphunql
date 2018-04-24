package gqlquery

import (
	"net/http"
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
}

// test Request
func TestClientRequest(t *testing.T) {
}

// test MultiRequest
func TestClientMultiRequest(t *testing.T) {
}

// a handler that always returns a hardcoded payload.
func fakeSuccessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/graphql")
	w.Write([]byte(`query(bar: "baz", quux: 1234) {
  first
  second
  third(arg1: "one", arg2: 2) {
    one
    two
  }
}`))
}
