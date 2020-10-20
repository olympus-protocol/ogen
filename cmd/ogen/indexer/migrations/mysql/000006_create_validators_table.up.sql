CREATE TABLE IF NOT EXISTS `validators` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `public_key` BINARY(48) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `id_UNIQUE` (`id` ASC) VISIBLE,
    UNIQUE INDEX `public_key_UNIQUE` (`public_key` ASC) VISIBLE,
    CONSTRAINT `data_public_key`
        FOREIGN KEY (`public_key`)
            REFERENCES `deposits` (`data_public_key`)
            ON DELETE NO ACTION
            ON UPDATE NO ACTION
);