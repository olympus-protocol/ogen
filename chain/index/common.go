package index

type Indexers struct {
	BlockIndex  *BlockIndex
	UtxoIndex   *UtxosIndex
	GovIndex    *GovIndex
	UserIndex   *UserIndex
	WorkerIndex *WorkerIndex
}
