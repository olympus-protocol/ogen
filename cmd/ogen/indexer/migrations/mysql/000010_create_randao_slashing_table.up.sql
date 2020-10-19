CREATE TABLE `randao_slashing` (
    `block_hash` BINARY(32) NOT NULL,
    `randao_reveal` BINARY(96) NOT NULL,
    `slot` INT NOT NULL,
    `validator_public_key` BINARY(48) NOT NULL,
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