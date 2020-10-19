CREATE TABLE IF NOT EXISTS `block_headers` (
    `id` INT NOT NULL,
    `block_hash` BINARY(32) NOT NULL,
    `version` INT NOT NULL,
    `nonce` INT NOT NULL,
    `tx_mekle_root` BINARY(32) NOT NULL,
    `tx_multi_merkle_root` BINARY(32) NOT NULL,
    `vote_merkle_root` BINARY(32) NOT NULL,
    `deposit_merkle_root` BINARY(32) NOT NULL,
    `exit_merkle_root` BINARY(32) NOT NULL,
    `vote_slashing_merkle_root` BINARY(32) NOT NULL,
    `randao_slashing_merkle_root` BINARY(32) NOT NULL,
    `proposer_slashing_merkle_root` BINARY(32) NOT NULL,
    `governance_votes_merkle_root` BINARY(32) NOT NULL,
    `previous_block_hash` BINARY(32) NOT NULL,
    `timestamp` INT NOT NULL,
    `slot` INT NOT NULL,
    `state_root` BINARY(32) NOT NULL,
    `fee_address` BINARY(20) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `block_hash_UNIQUE` (`block_hash` ASC) VISIBLE,
    CONSTRAINT `block_hash`
        FOREIGN KEY (`block_hash`)
            REFERENCES `blocks` (`block_hash`)
            ON DELETE NO ACTION
            ON UPDATE NO ACTION
);