#!/bin/bash

mockgen --source internal/chain/blockchain.go --destination internal/chain/blockchain_mock.go --package chain
mockgen --source internal/chain/state.go --destination internal/chain/state_mock.go --package chain
mockgen --source internal/mempool/votes.go --destination internal/mempool/votes_mock.go --package mempool
mockgen --source internal/peers/database.go --destination internal/peers/database_mock.go --package peers
mockgen --source internal/peers/discoveryprotocol.go --destination internal/peers/discoveryprotocol_mock.go --package peers
mockgen --source internal/peers/hostnode.go --destination internal/peers/hostnode_mock.go --package peers
mockgen --source internal/peers/protocolhandler.go --destination internal/peers/protocolhandler_mock.go --package peers
mockgen --source internal/peers/syncprotocol.go --destination internal/peers/syncprotocol_mock.go --package peers
mockgen --source internal/actionmanager/actionmanager.go --destination internal/actionmanager/actionmanager_mock.go --package actionmanager
mockgen --source internal/logger/log.go --destination internal/logger/log_mock.go --package logger
mockgen --source internal/state/interface.go --destination internal/state/interface_mock.go --package state