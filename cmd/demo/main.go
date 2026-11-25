package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"example.com/go-orm-vs-sql/internal/repository"
	"github.com/jackc/pgx/v5"
	"gorm.io/gorm"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

type Order struct {
	ID          int64
	UserID      int64
	Status      string
	TotalAmount int64
	CreatedAt   time.Time
}
type User struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
	Orders    []Order
}
type Product struct {
	ID    int64
	Name  string
	Price int64
	Stock int
}

func main() {
	mode := "crud"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	switch mode {
	case "-h", "--help", "help":
		fmt.Println("Usage: demo [crud|n-plus-one|zero-values|select-star|filters|transactions|bulk]")
		return
	case "crud", "n-plus-one", "zero-values", "select-star", "filters", "transactions", "bulk":
	default:
		log.Fatal("unknown demo: ", mode)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	c, err := repository.Open(ctx)
	check(err)
	defer c.Close()
	db := c.Gorm.WithContext(ctx)
	switch mode {
	case "crud":
		readers := c.Readers()
		for _, n := range []string{"pgx", "sqlx", "gorm", "sqlc"} {
			r := readers[n]
			u, e := r.GetUser(ctx, 1)
			check(e)
			fmt.Println(n, u)
		}
	case "n-plus-one":
		var users []User
		d := db.Debug()
		check(d.Limit(100).Order("id").Find(&users).Error)
		for i := range users {
			check(d.Where("user_id = ?", users[i].ID).Find(&users[i].Orders).Error)
		}
		fmt.Printf("%d queries; now preload\n", 1+len(users))
		users = nil
		check(d.Limit(100).Order("id").Preload("Orders").Find(&users).Error)
		preloadQueries := 1
		if len(users) > 0 {
			preloadQueries++
		}
		fmt.Printf("%d queries\n", preloadQueries)
	case "zero-values":
		d := db.Debug()
		check(d.Model(&Product{}).Where("id = ?", 1).Update("stock", 10).Error)
		check(d.Model(&Product{}).Where("id = ?", 1).Updates(Product{Stock: 0}).Error)
		var p Product
		check(d.First(&p, 1).Error)
		fmt.Println("struct:", p.Stock)
		check(d.Model(&p).Updates(map[string]any{"stock": 0}).Error)
		check(d.First(&p, 1).Error)
		fmt.Println("map:", p.Stock)
	case "select-star":
		for _, q := range []string{"SELECT * FROM users ORDER BY id LIMIT 1000", "SELECT id,name FROM users ORDER BY id LIMIT 1000"} {
			start := time.Now()
			rows, e := c.Pool.Query(ctx, q)
			check(e)
			n, payload := 0, 0
			for rows.Next() {
				n++
				for _, value := range rows.RawValues() {
					payload += len(value)
				}
			}
			check(rows.Err())
			rows.Close()
			fmt.Printf("%s: %d rows, %d payload bytes, %s\n", q, n, payload, time.Since(start))
		}
	case "filters":
		id := int64(1)
		q, args, e := repository.FilterOrders(&id, "completed")
		check(e)
		fmt.Println(q, args)
		rows, e := c.Pool.Query(ctx, q, args...)
		check(e)
		defer rows.Close()
		for rows.Next() {
			n, e := rows.Values()
			check(e)
			fmt.Println(n)
		}
		check(rows.Err())
	case "transactions":
		for _, safe := range []bool{false, true} {
			_, e := c.Pool.Exec(ctx, "UPDATE products SET stock=10 WHERE id=1")
			check(e)
			var wg sync.WaitGroup
			var successes atomic.Int64
			const workers = 100
			failures := make(chan error, workers)
			var reads sync.WaitGroup
			reads.Add(workers)
			for i := 0; i < workers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					var stock int
					var e error
					if safe {
						e = c.Pool.QueryRow(ctx, "UPDATE products SET stock=stock-1 WHERE id=1 AND stock>0 RETURNING stock").Scan(&stock)
						if errors.Is(e, pgx.ErrNoRows) {
							return
						}
					} else {
						e = c.Pool.QueryRow(ctx, "SELECT stock FROM products WHERE id=1").Scan(&stock)
						// Each SELECT releases its pooled connection before waiting.
						reads.Done()
						reads.Wait()
						if e == nil && stock > 0 {
							_, e = c.Pool.Exec(ctx, "UPDATE products SET stock=$1 WHERE id=1", stock-1)
						} else if e == nil {
							return
						}
					}
					if e != nil {
						failures <- e
						return
					}
					successes.Add(1)
				}()
			}
			wg.Wait()
			close(failures)
			for err := range failures {
				check(err)
			}
			var stock int
			check(c.Pool.QueryRow(ctx, "SELECT stock FROM products WHERE id=1").Scan(&stock))
			fmt.Println("atomic", safe, "success", successes.Load(), "stock", stock)
		}
	case "bulk":
		const n = 100000
		ps := make([]Product, n)
		rows := make([][]any, n)
		for i := range ps {
			ps[i] = Product{Name: "bulk", Price: 100, Stock: 1}
			rows[i] = []any{"bulk", 100, 1}
		}
		const setup = "CREATE TEMP TABLE bulk_products (LIKE products INCLUDING ALL) ON COMMIT DROP"
		check(db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(setup).Error; err != nil {
				return err
			}
			start := time.Now()
			result := tx.Table("bulk_products").CreateInBatches(ps, 1000)
			fmt.Println("gorm", result.RowsAffected, time.Since(start))
			return result.Error
		}))
		check(pgx.BeginFunc(ctx, c.Pool, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, setup); err != nil {
				return err
			}
			start := time.Now()
			count, err := tx.CopyFrom(ctx, pgx.Identifier{"bulk_products"}, []string{"name", "price", "stock"}, pgx.CopyFromRows(rows))
			fmt.Println("copy", count, time.Since(start))
			return err
		}))
	}
}
