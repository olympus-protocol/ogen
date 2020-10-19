CREATE TABLE IF NOT EXISTS `accounts` (
    `id` INT NOT NULL,
    `account` BINARY(20) NULL,
    `confirmed` INT NULL,
    `unconfirmed` INT NULL,
    `locked` INT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `id_UNIQUE` (`id` ASC) VISIBLE,
    UNIQUE INDEX `account_UNIQUE` (`account` ASC) VISIBLE
);