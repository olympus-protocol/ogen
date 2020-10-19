CREATE TABLE IF NOT EXISTS `blocks` (
    `block_hash` binary(32) NOT NULL,
    `block_signature` binary(96) NOT NULL,
    `block_randao_signature` binary(96) NOT NULL,
    `height` int NOT NULL AUTO_INCREMENT,
    PRIMARY KEY (`block_hash`),
    UNIQUE KEY `block_hash_UNIQUE` (`block_hash`),
    UNIQUE KEY `height_UNIQUE` (`height`)
)

