clean:
	rm -r *_encoding.go

build:
	sszgen -path ./msg_addr.go
	sszgen -path ./msg_block.go -include ../primitives/block.go,../primitives/blockheader.go,../primitives/votes.go,../primitives/tx.go,../primitives/deposit.go,../primitives/exit.go,../primitives/slashing.go,../primitives/governance.go
	sszgen -path ./msg_getaddr.go
	sszgen -path ./msg_getblocks.go
	sszgen -path ./msg_tx.go -include ../primitives/tx.go
	sszgen -path ./msg_version.go
