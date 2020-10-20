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