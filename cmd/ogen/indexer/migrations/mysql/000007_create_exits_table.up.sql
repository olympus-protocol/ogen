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