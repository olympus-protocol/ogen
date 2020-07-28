gen_ssz:
	sszgen -path ./validator.go
	sszgen -path ./votes.go
	sszgen -path ./deposit.go
	sszgen -path ./exit.go
	sszgen -path ./blockheader.go
	sszgen -path ./slashing.go -include ./votes.go,./blockheader.go
	# TODO TX
	sszgen -path ./tx.go