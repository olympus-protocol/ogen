create table blocks (
    block_hash text primary key unique not null,
    block_signature text not null,
    block_randao_signature text not null,
    height integer not null
);

