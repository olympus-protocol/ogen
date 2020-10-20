CREATE TABLE `randao_slashing` (
    `block_hash` binary(64) NOT NULL,
    `randao_reveal` binary(192) NOT NULL,
    `slot` INT NOT NULL,
    `validator_public_key` BINARY(96) NOT NULL,
    foreign key (block_hash) references blocks (block_hash)
);