CREATE TABLE IF NOT EXISTS `validators` (
    `id` INT AUTOINCREMENT,
    `public_key` BINARY(48) NOT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `data_public_key`
        FOREIGN KEY (`public_key`)
            REFERENCES `deposits` (`data_public_key`)
            ON DELETE NO ACTION
            ON UPDATE NO ACTION
);