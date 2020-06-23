clean:
	rm -r *_encoding.go

build:
	sszgen -path ./combined_signature.go
	sszgen -path ./multisig.go --include ../utils/bitfield