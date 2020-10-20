CREATE TABLE IF NOT EXISTS `exits` (
    `block_hash` binary(64) NOT NULL,
    `validator_public_key` binary(192) NOT NULL,
    `withdrawal_public_key` BINARY(20) NOT NULL,
    `signature` binary(192) NOT NULL,
    PRIMARY KEY (`block_hash`, `validator_public_key`),
    UNIQUE INDEX `block_hash_UNIQUE` (`block_hash` ASC) VISIBLE,
    UNIQUE INDEX `validator_public_key_UNIQUE` (`validator_public_key` ASC) VISIBLE,
    UNIQUE INDEX `signature_UNIQUE` (`signature` ASC) VISIBLE,
    FOREIGN KEY (`block_hash`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION,
    FOREIGN KEY (`validator_public_key`)
        REFERENCES `deposits` (`data_public_key`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);