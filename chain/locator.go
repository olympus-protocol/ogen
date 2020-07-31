package chain

// GetLocatorHashes for helping peers locate their.
func (ch *Blockchain) GetLocatorHashes() [64][32]byte {
	step := 1
	chain := ch.State().blockChain
	currentHeight := int64(chain.Tip().Height)
	locators := [64][32]byte{}
	for i := range locators {
		row, ok := chain.GetNodeByHeight(uint64(currentHeight))
		if !ok {
			break
		}
		currentHeight -= int64(step)
		step *= 2
		locators[i] = row.Hash.CloneBytes()
	}
	locators[63] = chain.Genesis().Hash.CloneBytes()
	return locators
}
