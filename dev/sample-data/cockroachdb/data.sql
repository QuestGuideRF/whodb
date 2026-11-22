CREATE DATABASE IF NOT EXISTS test_db;
CREATE USER IF NOT EXISTS "user";
GRANT admin TO "user";
USE test_db;
SET serial_normalization = 'sql_sequence';
CREATE SCHEMA IF NOT EXISTS test_schema;
GRANT ALL ON SCHEMA test_schema TO "user";
CREATE TABLE IF NOT EXISTS test_schema.users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL CHECK (LENGTH(password) >= 8),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS test_schema.products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    stock_quantity INT NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS test_schema.orders (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    status VARCHAR(20) CHECK (status IN ('pending', 'completed', 'canceled')) DEFAULT 'pending',
    FOREIGN KEY (user_id) REFERENCES test_schema.users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS test_schema.order_items (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL CHECK (quantity > 0),
    price_at_purchase DECIMAL(10,2) NOT NULL CHECK (price_at_purchase >= 0),
    FOREIGN KEY (order_id) REFERENCES test_schema.orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES test_schema.products(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS test_schema.payments (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL,
    payment_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    amount DECIMAL(10,2) NOT NULL CHECK (amount >= 0),
    payment_method VARCHAR(20) CHECK (payment_method IN ('credit_card', 'paypal', 'bank_transfer')),
    FOREIGN KEY (order_id) REFERENCES test_schema.orders(id) ON DELETE CASCADE
);
CREATE INDEX idx_users_email ON test_schema.users(email);
CREATE INDEX idx_orders_user_id ON test_schema.orders(user_id);
CREATE INDEX idx_order_items_order_id ON test_schema.order_items(order_id);
CREATE INDEX idx_payments_order_id ON test_schema.payments(order_id);
CREATE VIEW test_schema.order_summary AS
SELECT
    o.id AS order_id,
    u.username,
    o.order_date,
    o.status,
    o.total_amount
FROM test_schema.orders o
JOIN test_schema.users u ON o.user_id = u.id;
INSERT INTO test_schema.users (username, email, password) VALUES
('john_doe', 'john@example.com', 'securepassword1'),
('jane_smith', 'jane@example.com', 'securepassword2'),
('admin_user', 'admin@example.com', 'adminpass');
INSERT INTO test_schema.products (name, description, price, stock_quantity) VALUES
('Laptop', 'High-performance laptop', 1200.00, 10),
('Smartphone', 'Latest model smartphone', 800.00, 20),
('Headphones', 'Noise-canceling headphones', 150.00, 50),
('Monitor', '4K UHD Monitor', 400.00, 15);
INSERT INTO test_schema.orders (user_id, total_amount, status) VALUES
(1, 2000.00, 'completed'),
(2, 150.00, 'pending');
INSERT INTO test_schema.order_items (order_id, product_id, quantity, price_at_purchase) VALUES
(1, 1, 1, 1200.00),
(1, 2, 1, 800.00),
(2, 3, 1, 150.00);
INSERT INTO test_schema.payments (order_id, amount, payment_method) VALUES
(1, 2000.00, 'credit_card'),
(2, 150.00, 'paypal');
CREATE TABLE IF NOT EXISTS test_schema.test_casting (
    id SERIAL PRIMARY KEY,
    bigint_col BIGINT NOT NULL,
    integer_col INTEGER NOT NULL,
    smallint_col SMALLINT NOT NULL,
    numeric_col NUMERIC(10,2),
    description VARCHAR(100)
);
INSERT INTO test_schema.test_casting (bigint_col, integer_col, smallint_col, numeric_col, description)
VALUES (9223372036854775807, 2147483647, 32767, 99999999.99, 'Maximum values'),
       (1000000, 1000, 100, 1234.56, 'Regular values'),
       (-9223372036854775808, -2147483648, -32768, -99999999.99, 'Minimum values');
CREATE TABLE IF NOT EXISTS test_schema.data_types (
    id SERIAL PRIMARY KEY,
    col_smallint SMALLINT,
    col_integer INTEGER,
    col_bigint BIGINT,
    col_decimal DECIMAL(10,2),
    col_numeric NUMERIC(10,2),
    col_real REAL,
    col_double DOUBLE PRECISION,
    col_char CHAR(10),
    col_varchar VARCHAR(255),
    col_text TEXT,
    col_bytea BYTEA,
    col_timestamp TIMESTAMP,
    col_timestamptz TIMESTAMPTZ,
    col_date DATE,
    col_time TIME,
    col_boolean BOOLEAN,
    col_inet INET,
    col_uuid UUID,
    col_json JSON,
    col_jsonb JSONB
);
INSERT INTO test_schema.data_types (
    col_smallint, col_integer, col_bigint, col_decimal, col_numeric,
    col_real, col_double, col_char, col_varchar, col_text,
    col_bytea, col_timestamp, col_timestamptz, col_date, col_time,
    col_boolean, col_inet, col_uuid, col_json, col_jsonb
) VALUES (
    100, 1000, 100000, 123.45, 678.90,
    1.5, 2.5, 'test', 'varchar_val', 'text_value',
    E'\\x48454c4c4f', '2025-01-01 12:00:00', '2025-01-01 12:00:00+00', '2025-01-01', '12:00:00',
    true, '192.168.0.1',
    '550e8400-e29b-41d4-a716-446655440000',
    '{"key":"value"}', '{"key":"value"}'
);