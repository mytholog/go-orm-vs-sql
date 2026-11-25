package repository

import (
	"context"
	"example.com/go-orm-vs-sql/internal/dbgen"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time `db:"created_at"`
}
type Reader interface {
	GetUser(context.Context, int64) (User, error)
}

const query = `SELECT id,name,email,created_at FROM users WHERE id=$1`

type PGX struct{ Pool *pgxpool.Pool }

func (r PGX) GetUser(ctx context.Context, id int64) (u User, err error) {
	err = r.Pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	return
}

type SQLX struct{ DB *sqlx.DB }

func (r SQLX) GetUser(ctx context.Context, id int64) (u User, err error) {
	err = r.DB.GetContext(ctx, &u, query, id)
	return
}

type GORM struct{ DB *gorm.DB }

func (r GORM) GetUser(ctx context.Context, id int64) (u User, err error) {
	err = r.DB.WithContext(ctx).Select("id,name,email,created_at").Where("id = ?", id).Take(&u).Error
	return
}

type SQLC struct{ Q *dbgen.Queries }

func (r SQLC) GetUser(ctx context.Context, id int64) (User, error) {
	u, e := r.Q.GetUser(ctx, id)
	return User{u.ID, u.Name, u.Email, u.CreatedAt}, e
}
func FilterOrders(userID *int64, status string) (string, []any, error) {
	q := sq.Select("id", "user_id", "status", "total_amount").From("orders").PlaceholderFormat(sq.Dollar)
	if userID != nil {
		q = q.Where(sq.Eq{"user_id": *userID})
	}
	if status != "" {
		q = q.Where(sq.Eq{"status": status})
	}
	return q.OrderBy("id DESC").Limit(100).ToSql()
}
