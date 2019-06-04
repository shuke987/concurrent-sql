begin;
replace into t values (1,'2',3), (4,'5',6), (7,'8',9);
replace into t values (4,'5',6), (7,'8',9), (1,'2',3);
replace into t values (7,'8',9), (1,'2',3), (4,'5',6);
commit;