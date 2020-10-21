CREATE TABLE IF NOT EXISTS `blocks` (
    `block_hash` binary(64) NOT NULL,
    `block_signature` varbinary(192) NOT NULL,
    `block_randao_signature` varbinary(192) NOT NULL,
    `height` int NOT NULL,
    PRIMARY KEY (`block_hash`),
    UNIQUE KEY `block_hash_UNIQUE` (`block_hash`),
    UNIQUE KEY `height_UNIQUE` (`height`)
);

CREATE TABLE IF NOT EXISTS `block_headers` (
    `block_hash` binary(64) NOT NULL,
    `version` INT NOT NULL,
    `nonce` INT NOT NULL,
    `tx_merkle_root` binary(64) NOT NULL,
    `tx_multi_merkle_root` binary(64) NOT NULL,
    `vote_merkle_root` binary(64) NOT NULL,
    `deposit_merkle_root` binary(64) NOT NULL,
    `exit_merkle_root` binary(64) NOT NULL,
    `vote_slashing_merkle_root` binary(64) NOT NULL,
    `randao_slashing_merkle_root` binary(64) NOT NULL,
    `proposer_slashing_merkle_root` binary(64) NOT NULL,
    `governance_votes_merkle_root` binary(64) NOT NULL,
    `previous_block_hash` binary(64) NOT NULL,
    `timestamp` INT NOT NULL,
    `slot` INT NOT NULL,
    `state_root` binary(64) NOT NULL,
    `fee_address` binary(40) NOT NULL,
    UNIQUE INDEX `block_hash_UNIQUE` (`block_hash` ASC) VISIBLE,
    CONSTRAINT `block_hash`
        FOREIGN KEY (`block_hash`)
            REFERENCES `blocks` (`block_hash`)
            ON DELETE NO ACTION
            ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS `votes` (
    `block_hash` binary(64) NOT NULL,
    `signature` binary(192) NOT NULL,
    `participation_bitfield` varbinary(12516) NOT NULL,
    `data_slot` INT NOT NULL,
    `data_from_epoch` INT NOT NULL,
    `data_from_hash` binary(64) NOT NULL,
    `data_to_epoch` INT NOT NULL,
    `data_to_hash` binary(64) NOT NULL,
    `data_beacon_block_hash` binary(64) NOT NULL,
    `data_nonce` INT NOT NULL,
    `vote_hash` binary(64) NOT NULL,
    FOREIGN KEY (`block_hash`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS `deposits` (
    `block_hash` binary(64) NOT NULL,
    `public_key` BINARY(96) NOT NULL,
    `signature` binary(192) NOT NULL,
    `data_public_key` BINARY(96) NOT NULL,
    `data_proof_of_possession` binary(192) NOT NULL,
    `data_withdrawal_address` binary(40) NOT NULL,
    UNIQUE INDEX `data_public_key_UNIQUE` (`data_public_key` ASC) VISIBLE,
    UNIQUE INDEX `data_proof_of_possession_UNIQUE` (`data_proof_of_possession` ASC) VISIBLE,
    FOREIGN KEY (`block_hash`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS `accounts` (
    `account` binary(40) NOT NULL,
    `confirmed` INT DEFAULT 0,
    `unconfirmed` INT DEFAULT 0,
    `locked`INT DEFAULT 0,
    `total_sent` INT DEFAULT 0,
    `total_received`INT DEFAULT 0,
    UNIQUE INDEX `account_UNIQUE` (`account` ASC) VISIBLE
);

CREATE TABLE IF NOT EXISTS `validators` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `public_key` BINARY(96) NOT NULL,
    `exit` BOOLEAN NOT NULL,
    `penalized` BOOLEAN NOT NULL,
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `exits` (
    `block_hash` binary(64) NOT NULL,
    `validator_public_key` binary(192) NOT NULL,
    `withdrawal_public_key` binary(40) NOT NULL,
    `signature` binary(192) NOT NULL,
    FOREIGN KEY (`block_hash`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS `tx_single` (
    `block_hash` binary(64) NOT NULL,
    `tx_type` INT NOT NULL,
    `to_addr` BINARY(40) NOT NULL,
    `from_public_key` BINARY(96) NOT NULL,
    `amount` INT NOT NULL,
    `nonce` INT NOT NULL,
    `fee` INT NOT NULL,
    `signature` binary(192) NOT NULL,
    FOREIGN KEY (`block_hash`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS `vote_slashing` (
    `block_hash` binary(64) NOT NULL,
    `vote_1` binary(64) NOT NULL,
    `vote_2` binary(64) NOT NULL,
    PRIMARY KEY (`block_hash`, `vote_1`, `vote_2`),
    UNIQUE INDEX `block_hash_UNIQUE` (`block_hash` ASC) VISIBLE,
    FOREIGN KEY (`block_hash`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS `randao_slashing` (
    `block_hash` binary(64) NOT NULL,
    `randao_reveal` binary(192) NOT NULL,
    `slot` INT NOT NULL,
    `validator_public_key` BINARY(96) NOT NULL,
    PRIMARY KEY (`block_hash`),
    UNIQUE INDEX `block_hash_UNIQUE` (`block_hash` ASC) VISIBLE,
    FOREIGN KEY (`block_hash`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION,
    FOREIGN KEY (`validator_public_key`)
        REFERENCES `deposits` (`data_public_key`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);

CREATE TABLE IF NOT EXISTS `proposer_slashing` (
    `block_hash` binary(64) NOT NULL,
    `blockheader_1` binary(64) NOT NULL,
    `blockheader_2` binary(64) NOT NULL,
    `signature_1` binary(192) NOT NULL,
    `signature_2` binary(192) NOT NULL,
    `validator_public_key` BINARY(96) NOT NULL,
    PRIMARY KEY (`block_hash`),
    UNIQUE INDEX `block_hash_UNIQUE` (`block_hash` ASC) VISIBLE,
    FOREIGN KEY (`block_hash`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION,
    FOREIGN KEY (`blockheader_1`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION,
    FOREIGN KEY (`blockheader_2`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION,
    FOREIGN KEY (`validator_public_key`)
        REFERENCES `deposits` (`data_public_key`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);


