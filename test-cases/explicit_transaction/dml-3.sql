begin;
update ignore t set i = 3 where i = 6;
update ignore t set j = '3' where i = 3;
update ignore t set j = '4';
commit;