clean:
	rm -r *_encoding.go

build:
	sszgen -path ./block.go --include ./blockheader.go,./votes.go,./tx.go,./deposit.go,./exit.go,./slashing.go,./governance.go
	sszgen -path ./blockheader.go
	sszgen -path ./deposit.go
	sszgen -path ./exit.go
	sszgen -path ./governance.go
	sszgen -path ./validator.go
	sszgen -path ./coins_state.go
	sszgen -path ./slashing.go --include ./votes.go,./blockheader.go
	sszgen -path ./state.go --include ./validator.go,./coins_state.go,./votes.go
	sszgen -path ./tx.go --include ./validator.go 
	sszgen -path ./votes.go
