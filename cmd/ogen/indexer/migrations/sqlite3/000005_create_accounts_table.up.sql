create table accounts(
    account text primary key unique not null,
    confirmed integer default 0,
    unconfirmed integer default 0,
    locked integer default 0,
    total_sent integer default 0,
    total_received integer default 0
);