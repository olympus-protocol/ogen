#!/bin/bash

sszgen -path ./pkg/p2p/message.go -objs MessageHeader
sszgen -path ./pkg/p2p/msg_version.go
sszgen -path ./pkg/p2p/msg_finalization.go
sszgen -path ./pkg/p2p/msg_block.go -include ./pkg/primitives/block.go,./pkg/primitives/blockheader.go,./pkg/primitives/votes.go,./pkg/primitives/tx.go,./pkg/primitives/deposit.go,./pkg/primitives/exit.go,./pkg/primitives/slashing.go,./pkg/primitives/partialexit.go
sszgen -path ./pkg/p2p/msg_deposits.go -include ./pkg/primitives/deposit.go
sszgen -path ./pkg/p2p/msg_getblocks.go
sszgen -path ./pkg/p2p/msg_tx.go -include ./pkg/primitives/tx.go
sszgen -path ./pkg/p2p/msg_vote.go -include ./pkg/primitives/votes.go
sszgen -path ./pkg/p2p/msg_exits.go -include ./pkg/primitives/exit.go
sszgen -path ./pkg/p2p/msg_partialexit.go -include ./pkg/primitives/partialexit.go
sszgen -path ./pkg/primitives/block.go -include ./pkg/primitives/votes.go,./pkg/primitives/blockheader.go,./pkg/primitives/tx.go,./pkg/primitives/deposit.go,./pkg/primitives/exit.go,./pkg/primitives/slashing.go,./pkg/primitives/partialexit.go
sszgen -path ./pkg/primitives/blockheader.go
sszgen -path ./pkg/primitives/coins.go -objs CoinsStateSerializable
sszgen -path ./pkg/primitives/deposit.go
sszgen -path ./pkg/primitives/exit.go
sszgen -path ./pkg/primitives/partialexit.go
sszgen -path ./pkg/primitives/validator.go
sszgen -path ./pkg/primitives/votes.go
sszgen -path ./pkg/primitives/blockheader.go
sszgen -path ./pkg/primitives/slashing.go -include ./pkg/primitives/votes.go,./pkg/primitives/blockheader.go
sszgen -path ./pkg/primitives/tx.go
sszgen -path ./pkg/primitives/state.go -objs SerializableState -include ./pkg/primitives/coins.go,./pkg/primitives/validator.go,./pkg/primitives/votes.go
sszgen -path ./pkg/primitives/blocknodedisk.go
