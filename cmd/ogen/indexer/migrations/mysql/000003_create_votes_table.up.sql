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
    UNIQUE INDEX `block_hash_UNIQUE` (`block_hash` ASC) VISIBLE,
    UNIQUE INDEX `signature_UNIQUE` (`signature` ASC) VISIBLE,
    UNIQUE INDEX `vote_hash_UNIQUE` (`vote_hash` ASC) VISIBLE,
    FOREIGN KEY (`block_hash`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);