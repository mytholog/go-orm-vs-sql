//go:build integration

package benchmarks

import (
	"context"
	"testing"

	"example.com/go-orm-vs-sql/internal/repository"
)

func TestReaders(t *testing.T) {
	ctx := context.Background()
	c, e := repository.Open(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	want, e := c.Readers()["pgx"].GetUser(ctx, 1)
	if e != nil {
		t.Fatal(e)
	}
	for n, r := range c.Readers() {
		got, e := r.GetUser(ctx, 1)
		if e != nil || got.ID != want.ID || got.Name != want.Name || got.Email != want.Email || !got.CreatedAt.Equal(want.CreatedAt) {
			t.Errorf("%s: %+v %v", n, got, e)
		}
	}
}
