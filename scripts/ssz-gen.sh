#!/bin/bash

sszgen -path ./pkg/p2p/message.go
sszgen -path ./pkg/p2p/msg_version.go
sszgen -path ./pkg/p2p/msg_block.go -include ./pkg/primitives/block.go,./pkg/primitives/blockheader.go,./pkg/primitives/votes.go,./pkg/primitives/tx.go,./pkg/primitives/deposit.go,./pkg/primitives/exit.go,./pkg/primitives/slashing.go,./pkg/primitives/governance.go
sszgen -path ./pkg/p2p/msg_addr.go
sszgen -path ./pkg/p2p/msg_getblocks.go
sszgen -path ./pkg/primitives/block.go -include ./pkg/bls/multisig.go,./pkg/primitives/votes.go,./pkg/primitives/blockheader.go,./pkg/primitives/tx.go,./pkg/primitives/deposit.go,./pkg/primitives/exit.go,./pkg/primitives/slashing.go,./pkg/primitives/governance.go
sszgen -path ./pkg/primitives/blockheader.go
sszgen -path ./pkg/primitives/coins.go
sszgen -path ./pkg/primitives/deposit.go
sszgen -path ./pkg/primitives/exit.go
sszgen -path ./pkg/primitives/governance.go
sszgen -path ./pkg/primitives/validator.go
sszgen -path ./pkg/primitives/votes.go
sszgen -path ./pkg/primitives/blockheader.go
sszgen -path ./pkg/primitives/slashing.go -include ./pkg/primitives/votes.go,./pkg/primitives/blockheader.go
sszgen -path ./pkg/primitives/tx.go -include ./pkg/bls/multisig.go
sszgen -path ./pkg/primitives/state.go -objs SerializableState -include ./pkg/primitives/coins.go,./pkg/primitives/validator.go,./pkg/primitives/votes.go,./pkg/primitives/governance.go
sszgen -path ./pkg/bls/combined.go
sszgen -path ./pkg/bls/multisig.go
sszgen -path ./internal/chain/index/txs.go -objs TxLocator
sszgen -path ./internal/bdb/blocknodedisk.go
sszgen -path ./internal/peers/conflict/lastaction.go -objs ValidatorHelloMessage