CREATE TABLE IF NOT EXISTS `blocks` (
    `block_hash` binary(64) NOT NULL,
    `block_signature` varbinary(192) NOT NULL,
    `block_randao_signature` varbinary(192) NOT NULL,
    `height` int NOT NULL,
    PRIMARY KEY (`block_hash`),
    UNIQUE KEY `block_hash_UNIQUE` (`block_hash`),
    UNIQUE KEY `height_UNIQUE` (`height`)
);

