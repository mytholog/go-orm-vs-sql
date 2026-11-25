-- name: GetUser :one
SELECT id,name,email,created_at FROM users WHERE id=$1;
-- name: ReserveProduct :one
UPDATE products SET stock=stock-1 WHERE id=$1 AND stock>0 RETURNING stock;
