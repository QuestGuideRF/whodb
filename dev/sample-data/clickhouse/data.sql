DROP DATABASE IF EXISTS test_db;
CREATE DATABASE test_db;
CREATE TABLE IF NOT EXISTS test_db.users (
    id UInt32,
    username String,
    email String,
    password String,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY id;
CREATE TABLE IF NOT EXISTS test_db.products (
    id UInt32,
    name String,
    description String,
    price Decimal(10,2),
    stock_quantity UInt32 DEFAULT 0,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY id;
CREATE TABLE IF NOT EXISTS test_db.orders (
    id UInt32,
    user_id UInt32,
    order_date DateTime DEFAULT now(),
    total_amount Decimal(10,2) DEFAULT 0,
    status Enum('pending', 'completed', 'canceled') DEFAULT 'pending'
) ENGINE = MergeTree()
ORDER BY id;
CREATE TABLE IF NOT EXISTS test_db.order_items (
    id UInt32,
    order_id UInt32,
    product_id UInt32,
    quantity UInt32 DEFAULT 1,
    price_at_purchase Decimal(10,2)
) ENGINE = MergeTree()
ORDER BY id;
CREATE TABLE IF NOT EXISTS test_db.payments (
    id UInt32,
    order_id UInt32,
    payment_date DateTime DEFAULT now(),
    amount Decimal(10,2),
    payment_method Enum('credit_card', 'paypal', 'bank_transfer')
) ENGINE = MergeTree()
ORDER BY id;
CREATE TABLE IF NOT EXISTS test_db.test_casting (
    id INT,
    bigint_col BIGINT,
    integer_col INT,
    smallint_col SMALLINT,
    numeric_col DOUBLE,
    description TEXT
) ENGINE = MergeTree()
ORDER BY id;
CREATE MATERIALIZED VIEW IF NOT EXISTS test_db.order_summary
ENGINE = MergeTree()
ORDER BY order_id
POPULATE AS
SELECT
    o.id AS order_id,
    u.username AS username,
    o.order_date,
    o.status,
    o.total_amount
FROM test_db.orders o
INNER JOIN test_db.users u ON o.user_id = u.id;
INSERT INTO test_db.users (id, username, email, password, created_at) VALUES
(1, 'john_doe', 'john@example.com', 'securepassword1', now());
INSERT INTO test_db.users (id, username, email, password, created_at) VALUES
(2, 'jane_smith', 'jane@example.com', 'securepassword2', now());
INSERT INTO test_db.users (id, username, email, password, created_at) VALUES
(3, 'admin_user', 'admin@example.com', 'adminpass', now());
INSERT INTO test_db.products (id, name, description, price, stock_quantity, created_at) VALUES
(1, 'Laptop', 'High-performance laptop', 1200.00, 10, now());
INSERT INTO test_db.products (id, name, description, price, stock_quantity, created_at) VALUES
(2, 'Smartphone', 'Latest model smartphone', 800.00, 20, now());
INSERT INTO test_db.products (id, name, description, price, stock_quantity, created_at) VALUES
(3, 'Headphones', 'Noise-canceling headphones', 150.00, 50, now());
INSERT INTO test_db.products (id, name, description, price, stock_quantity, created_at) VALUES
(4, 'Monitor', '4K UHD Monitor', 400.00, 15, now());
INSERT INTO test_db.orders (id, user_id, order_date, total_amount, status) VALUES
(1, 1, now(), 2000.00, 'completed');
INSERT INTO test_db.orders (id, user_id, order_date, total_amount, status) VALUES
(2, 2, now(), 150.00, 'pending');
INSERT INTO test_db.order_items (id, order_id, product_id, quantity, price_at_purchase) VALUES
(1, 1, 1, 1, 1200.00);
INSERT INTO test_db.order_items (id, order_id, product_id, quantity, price_at_purchase) VALUES
(2, 1, 2, 1, 800.00);
INSERT INTO test_db.order_items (id, order_id, product_id, quantity, price_at_purchase) VALUES
(3, 2, 3, 1, 150.00);
INSERT INTO test_db.payments (id, order_id, payment_date, amount, payment_method) VALUES
(1, 1, now(), 2000.00, 'credit_card');
INSERT INTO test_db.payments (id, order_id, payment_date, amount, payment_method) VALUES
(2, 2, now(), 150.00, 'paypal');
INSERT INTO test_db.test_casting (id, bigint_col, integer_col, smallint_col, numeric_col, description) VALUES
(1, 9223372036854775807, 2147483647, 32767, 12345.67, 'Max values test');
INSERT INTO test_db.test_casting (id, bigint_col, integer_col, smallint_col, numeric_col, description) VALUES
(2, 1000000, 1000, 100, 99.99, 'Standard values');
INSERT INTO test_db.test_casting (id, bigint_col, integer_col, smallint_col, numeric_col, description) VALUES
(3, -1000000, -1000, -100, -99.99, 'Negative values');
CREATE TABLE IF NOT EXISTS test_db.data_types (
    id UInt32,
    col_int8 Int8,
    col_int16 Int16,
    col_int32 Int32,
    col_int64 Int64,
    col_float32 Float32,
    col_float64 Float64,
    col_decimal Decimal(10,2),
    col_decimal32 Decimal32(5),
    col_decimal64 Decimal64(3),
    col_date Date,
    col_datetime DateTime,
    col_string String,
    col_fixedstring FixedString(10),
    col_boolean UInt8,
    col_enum8 Enum8('active' = 1, 'inactive' = 2, 'pending' = 3),
    col_enum16 Enum16('small' = 1, 'medium' = 2, 'large' = 3, 'xlarge' = 4),
    col_nullable_int Nullable(Int32),
    col_nullable_string Nullable(String),
    col_lowcard LowCardinality(String),
    col_lowcard_nullable LowCardinality(Nullable(String)),
    col_array Array(Int32),
    col_array_string Array(String),
    col_map Map(String, Int32),
    col_tuple Tuple(String, Int32, Float64)
) ENGINE = MergeTree()
ORDER BY id;
INSERT INTO test_db.data_types (id, col_int8, col_int16, col_int32, col_int64, col_float32, col_float64, col_decimal, col_decimal32, col_decimal64, col_date, col_datetime, col_string, col_fixedstring, col_boolean, col_enum8, col_enum16, col_nullable_int, col_nullable_string, col_lowcard, col_lowcard_nullable, col_array, col_array_string, col_map, col_tuple) VALUES (1, 50, 1000, 100000, 10000000000, 1.5, 2.5, 123.45, 1234.56789, 123456789012345.678, '2025-01-01', '2025-01-01 12:00:00', 'string_val', 'fixed', 1, 'active', 'medium', 42, 'hello', 'category_a', 'tag_1', [10, 20, 30], ['foo', 'bar'], {'key1': 1, 'key2': 2}, ('label', 99, 3.14));
INSERT INTO test_db.data_types (id, col_int8, col_int16, col_int32, col_int64, col_float32, col_float64, col_decimal, col_decimal32, col_decimal64, col_date, col_datetime, col_string, col_fixedstring, col_boolean, col_enum8, col_enum16, col_nullable_int, col_nullable_string, col_lowcard, col_lowcard_nullable, col_array, col_array_string, col_map, col_tuple) VALUES (2, -10, -500, -50000, -5000000000, -0.5, -1.5, -67.89, -999.12345, -999.999, '2024-06-15', '2024-06-15 08:30:00', 'another_val', 'test', 0, 'inactive', 'large', NULL, NULL, 'category_b', NULL, [1, 2], ['baz'], {'only': 5}, ('other', -1, 0.0));