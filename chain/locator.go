package chain

import "github.com/olympus-protocol/ogen/utils/chainhash"

// GetLocatorHashes for helping peers locate their.
func (ch *Blockchain) GetLocatorHashes() []chainhash.Hash {
	step := 1
	chain := ch.State().blockChain
	currentHeight := int64(chain.Tip().Height)
	locators := []chainhash.Hash{}
	for currentHeight > 0 {
		row, ok := chain.GetNodeByHeight(uint64(currentHeight))
		if !ok {
			break
		}

		locators = append(locators, row.Hash)

		currentHeight -= int64(step)
		step *= 2
	}

	locators = append(locators, chain.Genesis().Hash)

	return locators
}
