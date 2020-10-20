CREATE TABLE `indexer`.`vote_slashing` (
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
