CREATE TABLE IF NOT EXISTS `validators` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `public_key` BINARY(96) NOT NULL,
    `exit` BOOLEAN NOT NULL,
    `penalized` BOOLEAN NOT NULL,
    PRIMARY KEY (`id`)
);