#!/bin/bash

rm -rf ./dataloader/*

go run github.com/vektah/dataloaden BlockSliceLoader string []*github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph.Block
go run github.com/vektah/dataloaden BlockLoader string *github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph.Block
go run github.com/vektah/dataloaden AccountsSliceLoader string []*github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph.Account
go run github.com/vektah/dataloaden AccountsLoader string *github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph.Account
go run github.com/vektah/dataloaden TxTxSingleSliceLoader string []*github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph.TxSingle
go run github.com/vektah/dataloaden TxTxSingleLoader string *github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph.TxSingle
go run github.com/vektah/dataloaden DepositSliceLoader string []*github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph.Deposit
go run github.com/vektah/dataloaden DepositLoader string *github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph.Deposit
go run github.com/vektah/dataloaden ExitSliceLoader string []*github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph.Exit
go run github.com/vektah/dataloaden ExitLoader string *github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph.Exit
go run github.com/vektah/dataloaden VotesSliceLoader string []*github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph.Votes
go run github.com/vektah/dataloaden VotesLoader string *github.com/olympus-protocol/ogen/cmd/ogen/indexer/graph.Votes

mv *_gen.go ./dataloader/