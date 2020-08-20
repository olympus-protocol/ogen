#!/bin/bash

sszgen -path ./pkg/p2p/message.go
sszgen -path ./pkg/p2p/msg_version.go
sszgen -path ./pkg/p2p/msg_blocks.go -include ./pkg/primitives/block.go,./pkg/primitives/blockheader.go,./pkg/primitives/votes.go,./pkg/primitives/tx.go,./pkg/primitives/tx_multi.go,./pkg/primitives/deposit.go,./pkg/primitives/exit.go,./pkg/primitives/slashing.go,./pkg/primitives/governance_votes.go,./pkg/bls/multisig/combined.go,./pkg/bls/multisig/multisig.go,./pkg/bls/multisig/combined.go
sszgen -path ./pkg/p2p/msg_addr.go
sszgen -path ./pkg/p2p/msg_getblocks.go
sszgen -path ./pkg/primitives/block.go -include ./pkg/bls/multisig/multisig.go,./pkg/primitives/votes.go,./pkg/primitives/blockheader.go,./pkg/primitives/tx.go,./pkg/primitives/tx_multi.go,./pkg/primitives/deposit.go,./pkg/primitives/exit.go,./pkg/primitives/slashing.go,./pkg/primitives/governance_votes.go,./pkg/bls/multisig/combined.go,./pkg/bls/multisig/multisig.go
sszgen -path ./pkg/primitives/blockheader.go
sszgen -path ./pkg/primitives/coins.go -objs CoinsStateSerializable
sszgen -path ./pkg/primitives/deposit.go
sszgen -path ./pkg/primitives/exit.go
sszgen -path ./pkg/primitives/governance.go -objs GovernanceSerializable -include ./pkg/primitives/governance_votes.go,./pkg/bls/multisig/combined.go,./pkg/bls/multisig/multisig.go
sszgen -path ./pkg/primitives/governance_votes.go -include ./pkg/bls/multisig/combined.go,./pkg/bls/multisig/multisig.go
sszgen -path ./pkg/primitives/validator.go
sszgen -path ./pkg/primitives/votes.go
sszgen -path ./pkg/primitives/blockheader.go
sszgen -path ./pkg/primitives/slashing.go -include ./pkg/primitives/votes.go,./pkg/primitives/blockheader.go
sszgen -path ./pkg/primitives/tx.go
sszgen -path ./pkg/primitives/tx_multi.go -include ./pkg/bls/multisig/multisig.go
sszgen -path ./internal/state/state.go -objs SerializableState -include ./pkg/primitives/coins.go,./pkg/primitives/validator.go,./pkg/primitives/votes.go,./pkg/primitives/governance.go,./pkg/primitives/governance_votes.go,./pkg/bls/multisig/combined.go,./pkg/bls/multisig/multisig.go
sszgen -path ./pkg/bls/multisig/combined.go
sszgen -path ./pkg/bls/multisig/multisig.go
sszgen -path ./internal/txindex/txlocator.go
sszgen -path ./internal/blockdb/blocknodedisk.go
sszgen -path ./internal/actionmanager/validatorhello.go