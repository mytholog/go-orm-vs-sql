package main

import (
	"context"
	"errors"
	"example.com/go-orm-vs-sql/internal/repository"
	"fmt"
	"github.com/jackc/pgx/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
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
	ctx := context.Background()
	c, e := repository.Open(ctx)
	check(e)
	defer c.Close()
	mode := "crud"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	db := c.Gorm.WithContext(ctx)
	switch mode {
	case "crud":
		for n, r := range c.Readers() {
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
		fmt.Println("101 queries; now preload")
		check(d.Limit(100).Order("id").Preload("Orders").Find(&users).Error)
		fmt.Println("2 queries")
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
		for _, q := range []string{"SELECT * FROM users LIMIT 1000", "SELECT id,name FROM users LIMIT 1000"} {
			start := time.Now()
			rows, e := c.Pool.Query(ctx, q)
			check(e)
			n := 0
			for rows.Next() {
				n++
			}
			check(rows.Err())
			rows.Close()
			fmt.Println(q, n, time.Since(start))
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
			var failures atomic.Int64
			for i := 0; i < 100; i++ {
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
						if e == nil && stock > 0 {
							time.Sleep(time.Millisecond)
							_, e = c.Pool.Exec(ctx, "UPDATE products SET stock=$1 WHERE id=1", stock-1)
						} else if e == nil {
							return
						}
					}
					if e != nil {
						failures.Add(1)
						return
					}
					successes.Add(1)
				}()
			}
			wg.Wait()
			var stock int
			check(c.Pool.QueryRow(ctx, "SELECT stock FROM products WHERE id=1").Scan(&stock))
			fmt.Println("atomic", safe, "success", successes.Load(), "errors", failures.Load(), "stock", stock)
		}
	case "bulk":
		const n = 100000
		ps := make([]Product, n)
		rows := make([][]any, n)
		for i := range ps {
			ps[i] = Product{Name: "bulk", Price: 100, Stock: 1}
			rows[i] = []any{"bulk", 100, 1}
		}
		for _, method := range []string{"gorm", "copy"} {
			_, e := c.Pool.Exec(ctx, "DELETE FROM products WHERE name='bulk'")
			check(e)
			start := time.Now()
			if method == "gorm" {
				check(db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).CreateInBatches(ps, 1000).Error)
			} else {
				count, e := c.Pool.CopyFrom(ctx, pgx.Identifier{"products"}, []string{"name", "price", "stock"}, pgx.CopyFromRows(rows))
				check(e)
				fmt.Println("copied", count)
			}
			fmt.Println(method, time.Since(start))
		}
	default:
		log.Fatal("unknown demo: ", mode)
	}
}
