-- Схема базы данных. Записи из рабочих таблиц исключены.

CREATE TABLE users (
    id INTEGER PRIMARY KEY DEFAULT nextval('users_id_seq'),
    username VARCHAR UNIQUE NOT NULL,
    email VARCHAR UNIQUE NOT NULL,
    password VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT current_timestamp,
    CHECK (LENGTH(password) >= 8)
);

CREATE TABLE products (
    id INTEGER PRIMARY KEY DEFAULT nextval('products_id_seq'),
    name VARCHAR NOT NULL UNIQUE,
    description VARCHAR,
    price DOUBLE NOT NULL CHECK (price >= 0),
    stock_quantity INTEGER NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    created_at TIMESTAMP DEFAULT current_timestamp
);

CREATE TABLE orders (
    id INTEGER PRIMARY KEY DEFAULT nextval('orders_id_seq'),
    user_id INTEGER NOT NULL,
    order_date TIMESTAMP DEFAULT current_timestamp,
    total_amount DOUBLE NOT NULL DEFAULT 0,
    status VARCHAR DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'canceled')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE NO ACTION
);

CREATE TABLE order_items (
    id INTEGER PRIMARY KEY DEFAULT nextval('order_items_id_seq'),
    order_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    price_at_purchase DOUBLE NOT NULL CHECK (price_at_purchase >= 0),
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE NO ACTION,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE NO ACTION
);

CREATE TABLE payments (
    id INTEGER PRIMARY KEY DEFAULT nextval('payments_id_seq'),
    order_id INTEGER NOT NULL,
    payment_date TIMESTAMP DEFAULT current_timestamp,
    amount DOUBLE NOT NULL CHECK (amount >= 0),
    payment_method VARCHAR CHECK (payment_method IN ('credit_card', 'paypal', 'bank_transfer')),
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE NO ACTION
);

CREATE INDEX idx_users_email ON users(email);

CREATE INDEX idx_orders_user_id ON orders(user_id);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);

CREATE INDEX idx_payments_order_id ON payments(order_id);

CREATE TABLE test_casting (
    id INTEGER PRIMARY KEY DEFAULT nextval('test_casting_id_seq'),
    bigint_col BIGINT NOT NULL,
    integer_col INTEGER NOT NULL,
    smallint_col SMALLINT NOT NULL,
    numeric_col DOUBLE,
    description VARCHAR
);

CREATE TABLE data_types (
    id INTEGER PRIMARY KEY DEFAULT nextval('data_types_id_seq'),
    col_tinyint TINYINT,
    col_smallint SMALLINT,
    col_integer INTEGER,
    col_bigint BIGINT,
    col_hugeint HUGEINT,
    col_utinyint UTINYINT,
    col_usmallint USMALLINT,
    col_uinteger UINTEGER,
    col_ubigint UBIGINT,
    col_float FLOAT,
    col_double DOUBLE,
    col_decimal DECIMAL(10,2),
    col_varchar VARCHAR,
    col_blob BLOB,
    col_boolean BOOLEAN,
    col_date DATE,
    col_time TIME,
    col_timestamp TIMESTAMP,
    col_timestamptz TIMESTAMP WITH TIME ZONE,
    col_interval INTERVAL,
    col_json JSON,
    col_uuid UUID
);
