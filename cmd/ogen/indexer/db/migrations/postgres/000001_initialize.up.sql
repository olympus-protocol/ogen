CREATE TABLE IF NOT EXISTS blocks (
    block_hash text NOT NULL,
    block_signature text NOT NULL,
    block_randao_signature text NOT NULL,
    height integer NOT NULL,
    UNIQUE(block_hash),
    PRIMARY KEY (block_hash)
);

CREATE TABLE IF NOT EXISTS block_headers (
    block_hash text NOT NULL,
    version BIGINT NOT NULL,
    nonce BIGINT NOT NULL,
    tx_merkle_root text NOT NULL,
    tx_multi_merkle_root text NOT NULL,
    vote_merkle_root text NOT NULL,
    deposit_merkle_root text NOT NULL,
    exit_merkle_root text NOT NULL,
    vote_slashing_merkle_root text NOT NULL,
    randao_slashing_merkle_root text NOT NULL,
    proposer_slashing_merkle_root text NOT NULL,
    governance_votes_merkle_root text NOT NULL,
    previous_block_hash text NOT NULL,
    timestamp BIGINT NOT NULL,
    slot BIGINT NOT NULL,
    state_root text NOT NULL,
    fee_address text NOT NULL,
    CONSTRAINT block_hash
        FOREIGN KEY (block_hash)
            REFERENCES blocks (block_hash)
            ON DELETE NO ACTION
            ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS votes (
    block_hash text NOT NULL,
    signature text NOT NULL,
    participation_bitfield text NOT NULL,
    data_slot BIGINT NOT NULL,
    data_from_epoch BIGINT NOT NULL,
    data_from_hash text NOT NULL,
    data_to_epoch BIGINT NOT NULL,
    data_to_hash text NOT NULL,
    data_beacon_block_hash text NOT NULL,
    data_nonce BIGINT NOT NULL,
    vote_hash text NOT NULL,
    FOREIGN KEY (block_hash)
        REFERENCES blocks (block_hash)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS deposits (
    block_hash text NOT NULL,
    public_key BYTEA NOT NULL,
    signature text NOT NULL,
    data_public_key BYTEA NOT NULL,
    data_proof_of_possession text NOT NULL,
    data_withdrawal_address text NOT NULL,
    UNIQUE(data_public_key),
    FOREIGN KEY (block_hash)
        REFERENCES blocks (block_hash)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS accounts (
    account text NOT NULL,
    confirmed BIGINT DEFAULT 0,
    unconfirmed BIGINT DEFAULT 0,
    locked BIGINT DEFAULT 0,
    total_sent BIGINT DEFAULT 0,
    total_received BIGINT DEFAULT 0
);

CREATE SEQUENCE validators_seq;

CREATE TABLE IF NOT EXISTS validators (
    id BIGINT NOT NULL DEFAULT NEXTVAL ('validators_seq'),
    public_key BYTEA NOT NULL,
    exit BOOLEAN NOT NULL,
    penalized BOOLEAN NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS exits (
    block_hash text NOT NULL,
    validator_public_key text NOT NULL,
    withdrawal_public_key text NOT NULL,
    signature text NOT NULL,
    FOREIGN KEY (block_hash)
        REFERENCES blocks (block_hash)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS tx_single (
    hash text NOT NULL,
    block_hash text NOT NULL,
    tx_type BIGINT NOT NULL,
    to_addr BYTEA NOT NULL,
    from_public_key BYTEA NOT NULL,
    amount BIGINT NOT NULL,
    nonce BIGINT NOT NULL,
    fee BIGINT NOT NULL,
    signature text NOT NULL,
    FOREIGN KEY (block_hash)
        REFERENCES blocks (block_hash)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS vote_slashing (
    block_hash text NOT NULL,
    vote_1 text NOT NULL,
    vote_2 text NOT NULL,
    PRIMARY KEY (block_hash, vote_1, vote_2),
    FOREIGN KEY (block_hash)
        REFERENCES blocks (block_hash)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS randao_slashing (
    block_hash text NOT NULL,
    randao_reveal text NOT NULL,
    slot BIGINT NOT NULL,
    validator_public_key BYTEA NOT NULL,
    PRIMARY KEY (block_hash),
    FOREIGN KEY (block_hash)
        REFERENCES blocks (block_hash)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION,
    FOREIGN KEY (validator_public_key)
        REFERENCES deposits (data_public_key)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS proposer_slashing (
    block_hash text NOT NULL,
    blockheader_1 text NOT NULL,
    blockheader_2 text NOT NULL,
    signature_1 text NOT NULL,
    signature_2 text NOT NULL,
    validator_public_key BYTEA NOT NULL,
    PRIMARY KEY (block_hash),
    FOREIGN KEY (block_hash)
        REFERENCES blocks (block_hash)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION,
    FOREIGN KEY (blockheader_1)
        REFERENCES blocks (block_hash)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION,
    FOREIGN KEY (blockheader_2)
        REFERENCES blocks (block_hash)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION,
    FOREIGN KEY (validator_public_key)
        REFERENCES deposits (data_public_key)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);


