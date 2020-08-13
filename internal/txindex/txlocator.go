package txindex

// TxLocator is a simple struct to find a database referenced to a block without building a full chainindex
type TxLocator struct {
	Hash  [32]byte `ssz-size:"32"`
	Block [32]byte `ssz-size:"32"`
	Index uint64
}

// Marshal encodes the data.
func (t *TxLocator) Marshal() ([]byte, error) {
	b, err := t.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxTxLocatorSize {
		return nil, ErrorCombinedSignatureSize
	}
	return b, nil
}

// Unmarshal decodes the data.
func (t *TxLocator) Unmarshal(b []byte) error {
	if len(b) > MaxTxLocatorSize {
		return ErrorCombinedSignatureSize
	}
	return t.UnmarshalSSZ(b)
}
