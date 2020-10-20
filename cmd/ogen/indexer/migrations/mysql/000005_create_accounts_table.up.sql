CREATE TABLE IF NOT EXISTS `accounts` (
    `account` binary(40) NOT NULL,
    `confirmed` INT DEFAULT 0,
    `unconfirmed` INT DEFAULT 0,
    `locked`INT DEFAULT 0,
    `total_sent` INT DEFAULT 0,
    `total_received`INT DEFAULT 0,
    UNIQUE INDEX `account_UNIQUE` (`account` ASC) VISIBLE
);