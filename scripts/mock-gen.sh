#!/bin/bash

rm -rf **/*_mock.go

mockgen --source internal/actionmanager/actionmanager.go --destination internal/actionmanager/actionmanager_mock.go --package actionmanager
mockgen --source internal/blockdb/interface.go --destination internal/blockdb/interface_mock.go --package blockdb
mockgen --source internal/chain/blockchain.go --destination internal/chain/blockchain_mock.go --package chain
mockgen --source internal/chain/state.go --destination internal/chain/state_mock.go --package chain
mockgen --source internal/chainrpc/server.go --destination internal/chainrpc/server_mock.go --package chainrpc
mockgen --source internal/keystore/keystore.go --destination internal/keystore/keystore_mock.go --package keystore
mockgen --source internal/logger/log.go --destination internal/logger/log_mock.go --package logger
mockgen --source internal/mempool/actions.go --destination internal/mempool/actions_mock.go --package mempool
mockgen --source internal/mempool/coins.go --destination internal/mempool/coins_mock.go --package mempool
mockgen --source internal/mempool/votes.go --destination internal/mempool/votes_mock.go --package mempool
mockgen --source internal/hostnode/database.go --destination internal/hostnode/database_mock.go --package hostnode
mockgen --source internal/hostnode/discoveryprotocol.go --destination internal/hostnode/discoveryprotocol_mock.go --package hostnode
mockgen --source internal/hostnode/hostnode.go --destination internal/hostnode/hostnode_mock.go --package hostnode
mockgen --source internal/hostnode/protocolhandler.go --destination internal/hostnode/protocolhandler_mock.go --package hostnode
mockgen --source internal/hostnode/syncprotocol.go --destination internal/hostnode/syncprotocol_mock.go --package hostnode
mockgen --source internal/proposer/proposer.go --destination internal/proposer/proposer_mock.go --package proposer
mockgen --source internal/server/server.go --destination internal/server/server_mock.go --package server
mockgen --source internal/state/interface.go --destination internal/state/interface_mock.go --package state
mockgen --source internal/wallet/wallet.go --destination internal/wallet/wallet_mock.go --package wallet
mockgen --source internal/execution/execution.go --destination internal/execution/execution_mock.go --package execution