package chain

// GetLocatorHashes for helping peers locate their.
func (ch *Blockchain) GetLocatorHashes() [][]byte {
	step := 1
	chain := ch.State().blockChain
	currentHeight := int64(chain.Tip().Height)
	locators := [][]byte{}
	for currentHeight > 0 {
		row, ok := chain.GetNodeByHeight(uint64(currentHeight))
		if !ok {
			break
		}
		if len(locators) > 64 {
			break
		}
		rowHash := row.Hash.CloneBytes()
		locators = append(locators, rowHash[:])

		currentHeight -= int64(step)
		step *= 2
	}
	genesisHash := chain.Genesis().Hash.CloneBytes()
	locators = append(locators, genesisHash[:])

	return locators
}
