DROP DATABASE IF EXISTS explicit_transaction;
CREATE DATABASE explicit_transaction;
USE explicit_transaction;
create table t (i int unsigned key, j varchar(20) unique key, k int);
set global tidb_disable_txn_auto_retry = off;
set global tidb_init_chunk_size = 1;
set global tidb_max_chunk_size = 32;
insert into t values (1,'2',3), (4,'5',6), (7,'8',9);
