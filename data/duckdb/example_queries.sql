INSTALL postgres;
ATTACH 'dbname=customer user=postgres password=postgres host=postgres' AS postgres_db (TYPE POSTGRES, READ_ONLY);
SHOW ALL TABLES;
SELECT * FROM postgres_db.account.accounts;

SELECT * FROM '/opt/data/customer.csv';
