create table votes (
    block_hash text not null,
    signature text not null,
    participation_bitfield text not null,
    data_slot integer not null,
    data_from_epoch integer not null,
    data_from_hash text not null,
    data_to_epoch integer not null,
    data_to_hash text not null,
    data_beacon_block_hash text not null,
    data_nonce integer not null,
    vote_hash text not null,
    foreign key (block_hash) references blocks (block_hash)
);
