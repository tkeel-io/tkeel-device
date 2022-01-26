package markdown

import (
	"testing"
)

func TestFilterSchema(t *testing.T) {
	s := "#/definitions/request.string"
	t.Log(FilterSchema(s))
}
