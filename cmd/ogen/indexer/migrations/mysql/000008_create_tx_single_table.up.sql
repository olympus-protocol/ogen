CREATE TABLE IF NOT EXISTS `tx_single` (
    `block_hash` binary(64) NOT NULL,
    `tx_type` INT NOT NULL,
    `to_addr` BINARY(40) NOT NULL,
    `from_public_key` BINARY(96) NOT NULL,
    `amount` INT NOT NULL,
    `nonce` INT NOT NULL,
    `fee` INT NOT NULL,
    `signature` binary(192) NOT NULL,
    FOREIGN KEY (`block_hash`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);