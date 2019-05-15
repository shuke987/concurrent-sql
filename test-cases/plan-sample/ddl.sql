DROP DATABASE IF EXISTS test2;
CREATE DATABASE test2;
USE test2;
CREATE TABLE `unknown_correlation` (id int, a int, PRIMARY KEY (`id`), INDEX idx_a(a));
