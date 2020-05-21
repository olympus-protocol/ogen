package chain_test

// var log = logger.New(os.Stdout).Quiet()

// const NumTestValidators = 128

// func getTestInitializationParameters() (*primitives.InitializationParameters, []bls.SecretKey) {
// 	vals := make([]primitives.ValidatorInitialization, NumTestValidators)
// 	keys := make([]bls.SecretKey, NumTestValidators)
// 	for i := range vals {
// 		k, err := bls.RandSecretKey(rand.Reader)
// 		if err != nil {
// 			panic(err)
// 		}

// 		keys[i] = *k

// 		vals[i] = primitives.ValidatorInitialization{
// 			PubKey:       keys[i].DerivePublicKey().Serialize(),
// 			PayeeAddress: "",
// 		}
// 	}

// 	return &primitives.InitializationParameters{
// 		InitialValidators: vals,
// 		GenesisTime:       time.Now().Add(1 * time.Second),
// 	}, keys
// }

// func TestBlockchainTipGenesis(t *testing.T) {
// 	db := mock.NewMemoryDB()

// 	ip, _ := getTestInitializationParameters()

// 	b, err := chain.NewBlockchain(chain.Config{
// 		Log: log,
// 	}, params.Mainnet, db, *ip)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	genesis := b.State().Tip()
// 	if genesis.Height != 0 {
// 		t.Fatal("expected genesis height to be 0")
// 	}

// 	if genesis.Parent != nil {
// 		t.Fatal("expected genesis parent to be nil")
// 	}
// }

// func TestBlockchainTipAddBlock(t *testing.T) {
// 	db := mock.NewMemoryDB()

// 	ip, _ := getTestInitializationParameters()

// 	b, err := chain.NewBlockchain(chain.Config{
// 		Log: log,
// 	}, params.Mainnet, db, *ip)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	genesis := b.State().Tip()
// 	if genesis.Height != 0 {
// 		t.Fatal("expected genesis height to be 0")
// 	}

// 	if genesis.Parent != nil {
// 		t.Fatal("expected genesis parent to be nil")
// 	}

// 	newBlockHeader := primitives.BlockHeader{
// 		Version:       0,
// 		Nonce:         0,
// 		MerkleRoot:    chainhash.Hash{},
// 		PrevBlockHash: genesis.Hash,
// 		Timestamp:     time.Time{},
// 	}

// 	msgHash := newBlockHeader.Hash()
// 	secretKey, _ := bls.RandSecretKey(rand.Reader)

// 	sig, err := bls.Sign(secretKey, msgHash[:])
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	err = b.ProcessBlock(&primitives.Block{
// 		Header:    newBlockHeader,
// 		Txs:       nil,
// 		Signature: sig.Serialize(),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }
