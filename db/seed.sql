TRUNCATE order_items, orders, products, users RESTART IDENTITY;
INSERT INTO users(name,email,profile) SELECT 'User '||i, i||'@example.test', jsonb_build_object('data',repeat(md5(i::text),128)) FROM generate_series(1,10000) i;
INSERT INTO products(name,price,stock) SELECT 'Product '||i, 1000, 10 FROM generate_series(1,1000) i;
INSERT INTO orders(user_id,status,total_amount,created_at) SELECT 1+(i%10000),'completed',1000,now()-(i%100000)*interval '1 minute' FROM generate_series(1,1000000) i;
INSERT INTO order_items(order_id,product_id,quantity,price) SELECT id,1+(id%1000),1,1000 FROM orders;
ANALYZE;
