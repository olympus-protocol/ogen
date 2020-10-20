create table exits (
    block_hash text not null,
    validator_public_key text not null,
    withdrawal_public_key text not null,
    signature text not null,
    foreign key (block_hash) references blocks (block_hash)
);