CREATE TABLE `proposer_slashing` (
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
