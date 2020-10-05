#!/bin/bash

rm -rf **/*_mock.go

mockgen --source internal/actionmanager/actionmanager.go --destination test/mocks/actionmanager/actionmanager.go --package actionmanager
mockgen --source internal/blockdb/interface.go --destination test/mocks/blockdb/interface.go --package blockdb
mockgen --source internal/chain/blockchain.go --destination test/mocks/chain/blockchain.go --package chain
mockgen --source internal/chain/state.go --destination test/mocks/chain/state.go --package chain
mockgen --source internal/chainrpc/server.go --destination test/mocks/chainrpc/server.go --package chainrpc
mockgen --source internal/keystore/keystore.go --destination test/mocks/keystore/keystore.go --package keystore
mockgen --source internal/mempool/actions.go --destination test/mocks/mempool/actions.go --package mempool
mockgen --source internal/mempool/coins.go --destination test/mocks/mempool/coins.go --package mempool
mockgen --source internal/mempool/votes.go --destination test/mocks/mempool/votes.go --package mempool
mockgen --source internal/hostnode/hostnode.go --destination test/mocks/hostnode/hostnode.go --package hostnode
mockgen --source internal/proposer/proposer.go --destination test/mocks/proposer/proposer.go --package proposer
mockgen --source internal/server/server.go --destination test/mocks/server/server.go --package server
mockgen --source internal/state/interface.go --destination test/mocks/state/interface.go --package state
mockgen --source internal/wallet/wallet.go --destination test/mocks/wallet/wallet.go --package wallet
mockgen --source internal/execution/execution.go --destination test/mocks/execution/execution.go --package execution