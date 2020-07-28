package chain

// GetLocatorHashes for helping peers locate their.
func (ch *Blockchain) GetLocatorHashes() [][32]byte {
	step := 1
	chain := ch.State().blockChain
	currentHeight := int64(chain.Tip().Height)
	locators := [][32]byte{}
	for currentHeight > 0 {
		row, ok := chain.GetNodeByHeight(uint64(currentHeight))
		if !ok {
			break
		}
		if len(locators) > 64 {
			break
		}
		locators = append(locators, row.Hash.CloneBytes())

		currentHeight -= int64(step)
		step *= 2
	}

	locators = append(locators, chain.Genesis().Hash.CloneBytes())

	return locators
}
