CREATE TABLE IF NOT EXISTS `blocks` (
    `block_hash` binary(64) NOT NULL UNIQUE,
    `block_signature` varbinary(192) NOT NULL,
    `block_randao_signature` varbinary(192) NOT NULL,
    `height` INT AUTOINCREMENT UNIQUE,
    PRIMARY KEY (`block_hash`)
)

