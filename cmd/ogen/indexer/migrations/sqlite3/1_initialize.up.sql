create table blocks (
                        block_hash text primary key unique not null,
                        block_signature text not null,
                        block_randao_signature text not null,
                        height integer not null
);

create table block_headers (
                               block_hash text unique not null,
                               version integer not null,
                               nonce integer not null,
                               tx_merkle_root text not null,
                               tx_multi_merkle_root text not null,
                               vote_merkle_root text not null,
                               deposit_merkle_root text not null,
                               exit_merkle_root text not null,
                               vote_slashing_merkle_root text not null,
                               randao_slashing_merkle_root text not null,
                               proposer_slashing_merkle_root text not null,
                               governance_votes_merkle_root text not null,
                               previous_block_hash text not null,
                               timestamp integer not null,
                               slot integer not null,
                               state_root text not null,
                               fee_address text not null,
                               foreign key (block_hash) references blocks (block_hash)
);

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

create table deposits (
                          block_hash text not null,
                          public_key text not null,
                          signature text not null,
                          data_public_key text not null,
                          data_proof_of_possession text not null,
                          data_withdrawal_address text not null,
                          foreign key (block_hash) references blocks (block_hash)
);

create table accounts(
                         account text primary key unique not null,
                         confirmed integer default 0,
                         unconfirmed integer default 0,
                         locked integer default 0,
                         total_sent integer default 0,
                         total_received integer default 0
);

create table validators (
                            id integer primary key autoincrement,
                            public_key text not null,
                            exit boolean not null,
                            penalized boolean not null
);

create table exits (
                       block_hash text not null,
                       validator_public_key text not null,
                       withdrawal_public_key text not null,
                       signature text not null,
                       foreign key (block_hash) references blocks (block_hash)
);

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

create table vote_slashing (
                               block_hash text not null,
                               vote_1 text not null,
                               vote_2 text not null,
                               foreign key (block_hash) references blocks (block_hash)
);

create table randao_slashings (
                                  block_hash text not null,
                                  randao_reveal text not null,
                                  slot integer not null,
                                  validator_public_key text not null,
                                  foreign key (block_hash) references blocks (block_hash)
);

create table proposer_slashing (
                                   block_hash text not null,
                                   blockheader_1 text not null,
                                   blockheader_2 text not null,
                                   signature_1  text not null,
                                   signature_2 text not null,
                                   validator_public_key  text not null,
                                   foreign key (block_hash) references blocks (block_hash)
);
