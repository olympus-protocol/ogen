CREATE TABLE IF NOT EXISTS `deposits` (
    `block_hash` binary(64) NOT NULL,
    `public_key` BINARY(48) NOT NULL,
    `signature` binary(192) NOT NULL,
    `data_public_key` BINARY(48) NOT NULL,
    `data_proof_of_possession` binary(192) NOT NULL,
    `data_withdrawal_address` BINARY(20) NOT NULL,
    PRIMARY KEY (`block_hash`),
    FOREIGN KEY (`block_hash`)
        REFERENCES `blocks` (`block_hash`)
        ON DELETE NO ACTION
        ON UPDATE NO ACTION
);