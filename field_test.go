package garphunql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldRender(t *testing.T) {
	f := Field("query",
		Arg("bar", "baz"),
		Arg("quux", 1234),
		GraphQLField{Name: "first", Alias: "foo"},
		Field("second"),
		Field("third",
			Alias("somealias"),
			Arg("arg1", "one"),
			Arg("arg2", 2),
			Field("one"),
			Field("two"),
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

func TestWrapFields(t *testing.T) {

	var d1 string
	var d2 string

	f1 := Field("user", Dest(&d1))
	f2 := Field("user", Dest(&d2))

	f, destMap := wrapFields("query", f1, f2)
	rendered, err := f.Render()
	assert.Nil(t, err)
	// assuming an 8 char random alias, the rendered result should look something like this.  But we
	// don't know what the random alias will be, and we don't know whether the aliased field will come
	// first or second.  A better test would parse the query and make assertions against the structure.
	assert.Equal(t, len("query {\n  nusghjbt: user\n  user\n}"), len(rendered))

	for k, _ := range destMap {
		assert.Contains(t, rendered, k)
	}
}
