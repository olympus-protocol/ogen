create table validators (
    id integer primary key autoincrement,
    public_key text not null,
    exit boolean not null,
    penalized boolean not null
);