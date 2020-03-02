package index

// Indexers contains all indexes needed for the blockchain.
type Indexers struct {
	BlockIndex  *BlockIndex
	UtxoIndex   *UtxosIndex
	GovIndex    *GovIndex
	UserIndex   *UserIndex
	WorkerIndex *WorkerIndex
}
