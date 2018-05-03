package garphunql

import "fmt"

// A GraphQLError represents an object returned in the "errors" list in a GraphQL response payload.
// It also implements the Go error interface.
type GraphQLError struct {
	Message   string                 `json:"message"`
	Locations []GraphQLErrorLocation `json:"locations"`
}

// Error renders the GraphQLError as a single string.
func (e GraphQLError) Error() string {
	return fmt.Sprintf("%s (%v)", e.Message, e.Locations)
}

// GraphQLErrorLocation is a sub-field of GraphQLError.
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}
