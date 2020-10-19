CREATE TABLE IF NOT EXISTS `votes` (
    `id` INT NOT NULL,
    `block_hash` BINARY(32) NOT NULL,
    `signature` BINARY(96) NOT NULL,
    `participation_bitfield` VARCHAR(6258) NOT NULL,
    `data_slot` INT NOT NULL,
    `data_from_epoch` INT NOT NULL,
    `data_from_hash` BINARY(32) NOT NULL,
    `data_to_epoch` INT NOT NULL,
    `data_to_hash` BINARY(32) NOT NULL,
    `data_beacon_block_hash` BINARY(32) NOT NULL,
    `data_nonce` INT NOT NULL,
    `vote_hash` BINARY(32) NOT NULL,
    PRIMARY KEY (`id`, `block_hash`),
    UNIQUE INDEX `block_hash_UNIQUE` (`block_hash` ASC) VISIBLE,
    UNIQUE INDEX `signature_UNIQUE` (`signature` ASC) VISIBLE,
    UNIQUE INDEX `vote_hash_UNIQUE` (`vote_hash` ASC) VISIBLE,
    FOREIGN KEY (`block_hash`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);