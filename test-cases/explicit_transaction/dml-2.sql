begin;
insert ignore into t values (1,'2',3), (4,'5',6), (7,'8',9) on duplicate key update j = i * k, i = values(i);
insert ignore into t values (1,'2',3), (4,'5',6), (7,'8',9) on duplicate key update j = i * k, i = values(i)+1;
insert ignore into t values (1,'2',3), (4,'5',6), (7,'8',9) on duplicate key update j = i * k, i = -values(i);
insert ignore into t values (1,'2',3), (4,'5',6), (7,'8',9) on duplicate key update j = i * k, i = -values(i)+3;
commit;