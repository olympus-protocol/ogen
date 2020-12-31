#!/bin/bash

./ogen --network testnet --dashboard & ./ogen indexer testnet --dbconn="postgresql://postgres@localhost:5432/indexer"
