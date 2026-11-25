package repository

import (
	"context"
	"example.com/go-orm-vs-sql/internal/dbgen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
)

type Connections struct {
	Pool *pgxpool.Pool
	SQL  *sqlx.DB
	Gorm *gorm.DB
}

func Open(ctx context.Context) (*Connections, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://demo:demo@localhost:5432/demo?sslmode=disable"
	}
	cfg, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		return nil, e
	}
	cfg.MaxConns = 10
	p, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		return nil, e
	}
	c, e := pgx.ParseConfig(dsn)
	if e != nil {
		p.Close()
		return nil, e
	}
	s := sqlx.NewDb(stdlib.OpenDB(*c), "pgx")
	s.SetMaxOpenConns(10)
	s.SetMaxIdleConns(10)
	g, e := gorm.Open(postgres.New(postgres.Config{Conn: s.DB}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if e != nil {
		s.Close()
		p.Close()
		return nil, e
	}
	if e = p.Ping(ctx); e != nil {
		s.Close()
		p.Close()
		return nil, e
	}
	return &Connections{p, s, g}, nil
}
func (c *Connections) Close() { c.SQL.Close(); c.Pool.Close() }
func (c *Connections) Readers() map[string]Reader {
	return map[string]Reader{"pgx": PGX{c.Pool}, "sqlx": SQLX{c.SQL}, "gorm": GORM{c.Gorm}, "sqlc": SQLC{dbgen.New(c.Pool)}}
}
