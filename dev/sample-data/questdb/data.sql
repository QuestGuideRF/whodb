-- Схема базы данных. Записи из рабочих таблиц исключены.

CREATE TABLE IF NOT EXISTS users (
    id INT,
    username STRING,
    email STRING,
    password STRING,
    created_at TIMESTAMP
) TIMESTAMP(created_at) PARTITION BY DAY;

CREATE TABLE IF NOT EXISTS products (
    id INT,
    name STRING,
    description STRING,
    price DOUBLE,
    stock_quantity INT,
    created_at TIMESTAMP
) TIMESTAMP(created_at) PARTITION BY DAY;

CREATE TABLE IF NOT EXISTS orders (
    id INT,
    user_id INT,
    total_amount DOUBLE,
    status STRING,
    order_date TIMESTAMP
) TIMESTAMP(order_date) PARTITION BY DAY;

CREATE TABLE IF NOT EXISTS order_items (
    id INT,
    order_id INT,
    product_id INT,
    quantity INT,
    price_at_purchase DOUBLE,
    created_at TIMESTAMP
) TIMESTAMP(created_at) PARTITION BY DAY;

CREATE TABLE IF NOT EXISTS payments (
    id INT,
    order_id INT,
    amount DOUBLE,
    payment_method STRING,
    payment_date TIMESTAMP
) TIMESTAMP(payment_date) PARTITION BY DAY;
