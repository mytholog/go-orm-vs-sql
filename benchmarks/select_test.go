package benchmarks

import (
	"context"
	"testing"

	"example.com/go-orm-vs-sql/internal/repository"
)

func BenchmarkGetUser(b *testing.B) {
	ctx := context.Background()
	c, e := repository.Open(ctx)
	if e != nil {
		b.Fatal(e)
	}
	defer c.Close()
	readers := c.Readers()
	for _, n := range []string{"pgx", "sqlx", "gorm", "sqlc"} {
		r := readers[n]
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
