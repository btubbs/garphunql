package garphunql

import (
	"encoding/json"
	"fmt"
)

// A GraphQLError represents an object returned in the "errors" list in a GraphQL response payload.
// It also implements the Go error interface.
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations"`
	Path       []string               `json:"path"`
	Extensions JSONMap                `json:"extensions"`
}

// Error renders the GraphQLError as a single string.
func (e GraphQLError) Error() string {
	newErr, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("%s", e.Message)
	}
	return fmt.Sprintf("%s", newErr)
}

// GraphQLErrorLocation is a sub-field of GraphQLError.
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}
