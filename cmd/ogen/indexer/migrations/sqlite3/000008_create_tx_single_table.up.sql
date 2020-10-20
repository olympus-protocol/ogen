create table tx_single (
    block_hash text not null,
    tx_type integer not null,
    to_addr text not null,
    from_public_key text not null,
    amount integer not null,
    nonce integer not null,
    fee integer not null,
    signature text not null,
    foreign key (block_hash) references blocks (block_hash)
);