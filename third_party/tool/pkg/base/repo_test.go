package base

import (
	"context"
	"testing"
)

func TestRepo(t *testing.T) {
	r := NewRepo("https://github.com/tkeel-io/tkeel-template-go.git", "")
	if err := r.Clone(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := r.CopyTo(context.Background(), "/tmp/test_repo", "github.com/tkeel-io/tkeel-template-go", nil); err != nil {
		t.Fatal(err)
	}
}
