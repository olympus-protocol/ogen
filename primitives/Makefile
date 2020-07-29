gen_ssz:
	sszgen -path ./block.go -include ./votes.go,./blockheader.go,./tx.go,./deposit.go,./exit.go,./slashing.go,./governance.go
	sszgen -path ./blockheader.go
	sszgen -path ./coins.go
	sszgen -path ./deposit.go
	sszgen -path ./exit.go
	sszgen -path ./governance.go
	sszgen -path ./validator.go
	sszgen -path ./votes.go
	sszgen -path ./blockheader.go
	sszgen -path ./slashing.go -include ./votes.go,./blockheader.go
	sszgen -path ./tx.go