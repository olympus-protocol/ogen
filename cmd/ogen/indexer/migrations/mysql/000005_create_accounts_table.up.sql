CREATE TABLE IF NOT EXISTS `accounts` (
    `account` BINARY(20) NOT NULL,
    `confirmed` INT NULL,
    `unconfirmed` INT NULL,
    `locked` INT NULL,
    UNIQUE INDEX `account_UNIQUE` (`account` ASC) VISIBLE
);