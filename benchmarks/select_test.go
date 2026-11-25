package benchmarks

import (
	"context"
	"example.com/go-orm-vs-sql/internal/repository"
	"testing"
)

func BenchmarkGetUser(b *testing.B) {
	ctx := context.Background()
	c, e := repository.Open(ctx)
	if e != nil {
		b.Fatal(e)
	}
	defer c.Close()
	for n, r := range c.Readers() {
		b.Run(n, func(b *testing.B) {
			if _, e := r.GetUser(ctx, 1); e != nil {
				b.Fatal(e)
			}
			b.ReportAllocs()
			for b.Loop() {
				if _, e := r.GetUser(ctx, 1); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
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
