create table vote_slashing (
    block_hash text not null,
    vote_1 text not null,
    vote_2 text not null,
    foreign key (block_hash) references blocks (block_hash)
);
