begin;
delete a from t as a where i = 6;
delete a from t as a;
commit;