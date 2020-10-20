create table deposits (
    block_hash text not null,
    public_key text not null,
    signature text not null,
    data_public_key text not null,
    data_proof_of_possession text not null,
    data_withdrawal_address text not null,
    foreign key (block_hash) references blocks (block_hash)
)
