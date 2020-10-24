CREATE TABLE `blocks` (
  `block_hash` varchar(255) PRIMARY KEY NOT NULL,
  `block_signature` varchar(255) NOT NULL,
  `block_randao_signature` varchar(255) NOT NULL,
  `height` int NOT NULL
);

CREATE TABLE `block_headers` (
  `block_hash` varchar(255) NOT NULL,
  `version` int NOT NULL,
  `nonce` bigint NOT NULL,
  `tx_merkle_root` varchar(255) NOT NULL,
  `tx_multi_merkle_root` varchar(255) NOT NULL,
  `vote_merkle_root` varchar(255) NOT NULL,
  `deposit_merkle_root` varchar(255) NOT NULL,
  `exit_merkle_root` varchar(255) NOT NULL,
  `vote_slashing_merkle_root` varchar(255) NOT NULL,
  `randao_slashing_merkle_root` varchar(255) NOT NULL,
  `proposer_slashing_merkle_root` varchar(255) NOT NULL,
  `governance_votes_merkle_root` varchar(255) NOT NULL,
  `previous_block_hash` varchar(255) NOT NULL,
  `timestamp` int NOT NULL,
  `slot` int NOT NULL,
  `state_root` varchar(255) NOT NULL,
  `fee_address` varchar(255) NOT NULL
);

CREATE TABLE `votes` (
  `block_hash` varchar(255) NOT NULL,
  `signature` varchar(255) NOT NULL,
  `participation_bitfield` varchar(12518) NOT NULL,
  `data_slot` int NOT NULL,
  `data_from_epoch` int NOT NULL,
  `data_from_hash` varchar(255) NOT NULL,
  `data_to_epoch` int NOT NULL,
  `data_to_hash` varchar(255) NOT NULL,
  `data_beacon_block_hash` varchar(255) NOT NULL,
  `data_nonce` bigint NOT NULL,
  `vote_hash` varchar(255) NOT NULL
);

CREATE TABLE `deposits` (
  `block_hash` varchar(255) NOT NULL,
  `public_key` varchar(255) NOT NULL,
  `signature` varchar(255) NOT NULL,
  `data_public_key` varchar(255) NOT NULL,
  `data_proof_of_possession` varchar(255) NOT NULL,
  `data_withdrawal_address` varchar(255) NOT NULL
);

CREATE TABLE `accounts` (
  `account` varchar(255) NOT NULL,
  `confirmed` bigint DEFAULT 0,
  `unconfirmed` bigint DEFAULT 0,
  `locked` bigint DEFAULT 0,
  `total_sent` bigint DEFAULT 0,
  `total_received` bigint DEFAULT 0
);

CREATE TABLE `validators` (
  `id` int PRIMARY KEY NOT NULL AUTO_INCREMENT,
  `public_key` varchar(255) NOT NULL,
  `status` int NOT NULL,
  `exit` boolean NOT NULL,
  `penalized` boolean NOT NULL,
  `balance` bigint DEFAULT 0
);

CREATE TABLE `exits` (
  `block_hash` varchar(255) NOT NULL,
  `validator_public_key` varchar(255) NOT NULL,
  `withdrawal_public_key` varchar(255) NOT NULL,
  `signature` varchar(255) NOT NULL
);

CREATE TABLE `tx_single` (
  `hash` varchar(255) NOT NULL,
  `block_hash` varchar(255) NOT NULL,
  `tx_type` int NOT NULL,
  `to_addr` varchar(255) NOT NULL,
  `from_public_key` varchar(255) NOT NULL,
  `from_public_key_hash` varchar(255) NOT NULL,
  `amount` bigint NOT NULL,
  `nonce` int NOT NULL,
  `fee` bigint NOT NULL,
  `signature` varchar(255) NOT NULL
);

CREATE TABLE `vote_slashing` (
  `block_hash` varchar(255) NOT NULL,
  `vote_1` varchar(255) NOT NULL,
  `vote_2` varchar(255) NOT NULL
);

CREATE TABLE `randao_slashing` (
  `block_hash` varchar(255) PRIMARY KEY NOT NULL,
  `randao_reveal` varchar(255) NOT NULL,
  `slot` int NOT NULL,
  `validator_public_key` varchar(255) NOT NULL
);

CREATE TABLE `proposer_slashing` (
  `block_hash` varchar(255) PRIMARY KEY NOT NULL,
  `blockheader_1` varchar(255) NOT NULL,
  `blockheader_2` varchar(255) NOT NULL,
  `signature_1` varchar(255) NOT NULL,
  `signature_2` varchar(255) NOT NULL,
  `validator_public_key` varchar(255) NOT NULL
);

CREATE TABLE `slots` (
  `slot` int PRIMARY KEY NOT NULL,
  `block_hash` varchar(255) NOT NULL,
  `committee` varchar(12518) NOT NULL,
  `proposer_index` int NOT NULL,
  `proposed` boolean NOT NULL,
  `participation_percentage` int NOT NULL
);

CREATE TABLE `epochs` (
  `epoch` int PRIMARY KEY NOT NULL,
  `slot_1` int NOT NULL,
  `slot_2` int NOT NULL,
  `slot_3` int NOT NULL,
  `slot_4` int NOT NULL,
  `slot_5` int NOT NULL,
  `participation_percentage` int NOT NULL
);

ALTER TABLE `blocks` ADD FOREIGN KEY (`block_hash`) REFERENCES `block_headers` (`block_hash`);

ALTER TABLE `block_headers` ADD FOREIGN KEY (`block_hash`) REFERENCES `votes` (`block_hash`);

ALTER TABLE `block_headers` ADD FOREIGN KEY (`block_hash`) REFERENCES `deposits` (`block_hash`);

ALTER TABLE `block_headers` ADD FOREIGN KEY (`block_hash`) REFERENCES `exits` (`block_hash`);

ALTER TABLE `block_headers` ADD FOREIGN KEY (`block_hash`) REFERENCES `vote_slashing` (`block_hash`);

ALTER TABLE `block_headers` ADD FOREIGN KEY (`block_hash`) REFERENCES `randao_slashing` (`block_hash`);

ALTER TABLE `block_headers` ADD FOREIGN KEY (`block_hash`) REFERENCES `proposer_slashing` (`block_hash`);

ALTER TABLE `block_headers` ADD FOREIGN KEY (`block_hash`) REFERENCES `proposer_slashing` (`blockheader_1`);

ALTER TABLE `block_headers` ADD FOREIGN KEY (`block_hash`) REFERENCES `proposer_slashing` (`blockheader_2`);

ALTER TABLE `block_headers` ADD FOREIGN KEY (`block_hash`) REFERENCES `tx_single` (`block_hash`);

ALTER TABLE `deposits` ADD FOREIGN KEY (`data_public_key`) REFERENCES `validators` (`public_key`);

ALTER TABLE `deposits` ADD FOREIGN KEY (`data_public_key`) REFERENCES `proposer_slashing` (`validator_public_key`);

ALTER TABLE `deposits` ADD FOREIGN KEY (`public_key`) REFERENCES `exits` (`validator_public_key`);

ALTER TABLE `deposits` ADD FOREIGN KEY (`public_key`) REFERENCES `randao_slashing` (`validator_public_key`);

ALTER TABLE `accounts` ADD FOREIGN KEY (`account`) REFERENCES `tx_single` (`to_addr`);

ALTER TABLE `accounts` ADD FOREIGN KEY (`account`) REFERENCES `tx_single` (`from_public_key_hash`);

ALTER TABLE `slots` ADD FOREIGN KEY (`slot`) REFERENCES `epochs` (`slot_1`);

ALTER TABLE `slots` ADD FOREIGN KEY (`slot`) REFERENCES `epochs` (`slot_2`);

ALTER TABLE `slots` ADD FOREIGN KEY (`slot`) REFERENCES `epochs` (`slot_3`);

ALTER TABLE `slots` ADD FOREIGN KEY (`slot`) REFERENCES `epochs` (`slot_4`);

ALTER TABLE `slots` ADD FOREIGN KEY (`slot`) REFERENCES `epochs` (`slot_5`);
