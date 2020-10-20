create table proposer_slashing (
    block_hash text not null,
    blockheader_1 text not null,
    blockheader_2 text not null,
    signature_1  text not null,
    signature_2 text not null,
    validator_public_key  text not null,
    foreign key (block_hash) references blocks (block_hash)
);
