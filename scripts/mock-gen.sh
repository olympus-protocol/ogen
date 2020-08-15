#!/bin/bash

rm -rf **/*_mock.go

mockgen --source internal/actionmanager/actionmanager.go --destination internal/actionmanager/actionmanager_mock.go --package actionmanager
mockgen --source internal/blockdb/blockdb.go --destination internal/blockdb/blockdb_mock.go --package blockdb
mockgen --source internal/chain/blockchain.go --destination internal/chain/blockchain_mock.go --package chain
mockgen --source internal/chain/state.go --destination internal/chain/state_mock.go --package chain
mockgen --source internal/chainrpc/server.go --destination internal/chainrpc/server_mock.go --package chainrpc
mockgen --source internal/keystore/keystore.go --destination internal/keystore/keystore_mock.go --package keystore
mockgen --source internal/logger/log.go --destination internal/logger/log_mock.go --package logger
mockgen --source internal/mempool/actions.go --destination internal/mempool/actions_mock.go --package mempool
mockgen --source internal/mempool/coins.go --destination internal/mempool/coins_mock.go --package mempool
mockgen --source internal/mempool/votes.go --destination internal/mempool/votes_mock.go --package mempool
mockgen --source internal/peers/database.go --destination internal/peers/database_mock.go --package peers
mockgen --source internal/peers/discoveryprotocol.go --destination internal/peers/discoveryprotocol_mock.go --package peers
mockgen --source internal/peers/hostnode.go --destination internal/peers/hostnode_mock.go --package peers
mockgen --source internal/peers/protocolhandler.go --destination internal/peers/protocolhandler_mock.go --package peers
mockgen --source internal/peers/syncprotocol.go --destination internal/peers/syncprotocol_mock.go --package peers
mockgen --source internal/proposer/proposer.go --destination internal/proposer/proposer_mock.go --package proposer
mockgen --source internal/server/server.go --destination internal/server/server_mock.go --package server
mockgen --source internal/state/interface.go --destination internal/state/interface_mock.go --package state
mockgen --source internal/txindex/txindex.go --destination internal/txindex/txindex_mock.go --package txindex
mockgen --source internal/wallet/wallet.go --destination internal/wallet/wallet_mock.go --package wallet