package models

type Block struct {
	Header          BlockHeader
	Signature       string
	RandaoSignature string
	Height          int
}
