create
extension pg_trgm;
/* trigram extension for faster text-search */

drop table if exists val_stats;
create table val_stats
(
    operator_addr   text   not null,
    start_block     bigint not null default 0,
    end_block       bigint not null default 0,
    sign_num        int    not null default 0,
    missed_sign_num int    not null default 0,
    uptime          float  not null default 0,
    primary key (operator_addr, start_block, end_block)
);

drop table if exists val_sign_p;
create table val_sign_p
(
    operator_addr text    not null,
    block_height  bigint  not null default 0,
    status        int     not null, /* Can be 0 = scheduled, 1 executed, 2 missed */
    child_table   int     not null,
    primary key (operator_addr, block_height, child_table)
) PARTITION BY LIST (child_table);

CREATE TABLE val_sign_0 PARTITION OF val_sign_p FOR VALUES IN
(
    0
);
CREATE TABLE val_sign_1 PARTITION OF val_sign_p FOR VALUES IN
(
    1
);
CREATE TABLE val_sign_2 PARTITION OF val_sign_p FOR VALUES IN
(
    2
);
CREATE TABLE val_sign_3 PARTITION OF val_sign_p FOR VALUES IN
(
    3
);
CREATE TABLE val_sign_4 PARTITION OF val_sign_p FOR VALUES IN
(
    4
);
CREATE TABLE val_sign_5 PARTITION OF val_sign_p FOR VALUES IN
(
    5
);
CREATE TABLE val_sign_6 PARTITION OF val_sign_p FOR VALUES IN
(
    6
);
CREATE TABLE val_sign_7 PARTITION OF val_sign_p FOR VALUES IN
(
    7
);
CREATE TABLE val_sign_8 PARTITION OF val_sign_p FOR VALUES IN
(
    8
);
CREATE TABLE val_sign_9 PARTITION OF val_sign_p FOR VALUES IN
(
    9
);

drop table if exists val_sign_missed;
create table val_sign_missed
(
    operator_addr text   not null,
    block_height  bigint not null default 0,
    primary key (operator_addr, block_height)
);


