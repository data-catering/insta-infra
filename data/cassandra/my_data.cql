CREATE KEYSPACE IF NOT EXISTS account WITH REPLICATION = {
   'class' : 'SimpleStrategy',
   'replication_factor' : 1
};
CREATE TABLE IF NOT EXISTS account.accounts(
    account_id TEXT,
    amount DOUBLE,
    created_by TEXT,
    name TEXT,
    open_time TIMESTAMP,
    status TEXT,
    PRIMARY KEY (account_id)
);
CREATE TABLE IF NOT EXISTS account.account_status_history(
    account_id TEXT,
    eod_date DATE,
    status TEXT,
    updated_by TEXT,
    updated_time TIMESTAMP,
    PRIMARY KEY ((account_id), eod_date)
);
