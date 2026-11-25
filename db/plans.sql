DROP INDEX IF EXISTS idx_orders_user_created_at;
EXPLAIN (ANALYZE, BUFFERS) SELECT id,status,total_amount FROM orders WHERE user_id=100 ORDER BY created_at DESC LIMIT 20;
CREATE INDEX idx_orders_user_created_at ON orders(user_id,created_at DESC);
EXPLAIN (ANALYZE, BUFFERS) SELECT id,status,total_amount FROM orders WHERE user_id=100 ORDER BY created_at DESC LIMIT 20;
SELECT user_id,SUM(total_amount) AS revenue FROM orders WHERE status='completed' AND created_at>=now()-interval '30 days' GROUP BY user_id ORDER BY revenue DESC LIMIT 5;
