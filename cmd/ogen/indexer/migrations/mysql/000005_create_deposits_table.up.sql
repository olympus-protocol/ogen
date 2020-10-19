CREATE TABLE IF NOT EXISTS `deposits` (
    `block_hash` BINARY(32) NOT NULL,
    `public_key` BINARY(48) NOT NULL,
    `signature` BINARY(96) NOT NULL,
    `data_public_key` BINARY(48) NOT NULL,
    `data_proof_of_possession` BINARY(96) NOT NULL,
    `data_withdrawal_address` BINARY(20) NOT NULL,
    PRIMARY KEY (`block_hash`),
    UNIQUE INDEX `block_hash_UNIQUE` (`block_hash` ASC) VISIBLE,
    UNIQUE INDEX `data_public_key_UNIQUE` (`data_public_key` ASC) VISIBLE,
    UNIQUE INDEX `data_proof_of_possession_UNIQUE` (`data_proof_of_possession` ASC) VISIBLE,
    CONSTRAINT `block_hash`
        FOREIGN KEY (`block_hash`)
            REFERENCES `blocks` (`block_hash`)
            ON DELETE NO ACTION
            ON UPDATE NO ACTION
);