CREATE DATABASE customer;

CREATE TABLE IF NOT EXISTS customer.accounts
(
    id                      SERIAL PRIMARY KEY,
    account_number          VARCHAR(20) NOT NULL,
    account_status          VARCHAR(10),
    created_by              TEXT,
    created_by_fixed_length CHAR(10),
    customer_id_int         INT,
    customer_id_smallint    SMALLINT,
    customer_id_bigint      BIGINT,
    customer_id_decimal     DECIMAL,
    customer_id_real        REAL,
    customer_id_double      DOUBLE PRECISION,
    open_date               DATE,
    open_timestamp          TIMESTAMP,
    last_opened_time        TIME,
    payload_bytes           BLOB
);
-- spark converts to wrong data type when reading from postgres so fails to write back to postgres
--     open_date_interval INTERVAL,
-- ERROR: column "open_date_interval" is of type interval but expression is of type character varying
--     open_id UUID,
--     balance MONEY,
--     payload_json JSONB

CREATE TABLE IF NOT EXISTS customer.balances
(
    id          BIGINT UNSIGNED NOT NULL,
    create_time TIMESTAMP,
    balance     DOUBLE PRECISION,
    PRIMARY KEY (id, create_time),
    CONSTRAINT fk_bal_account_number FOREIGN KEY (id) REFERENCES customer.accounts (id)
);

CREATE TABLE IF NOT EXISTS customer.transactions
(
    id             BIGINT UNSIGNED NOT NULL,
    create_time    TIMESTAMP,
    transaction_id VARCHAR(20),
    amount         DOUBLE PRECISION,
    PRIMARY KEY (id, create_time, transaction_id),
    CONSTRAINT fk_txn_account_number FOREIGN KEY (id) REFERENCES customer.accounts (id)
);
