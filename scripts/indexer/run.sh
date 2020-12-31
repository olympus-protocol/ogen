#!/bin/bash

./ogen --network testnet --dashboard & ./ogen indexer testnet --dbconn="postgresql://localhost:5432/indexer"
